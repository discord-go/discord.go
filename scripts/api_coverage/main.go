// Command api_coverage reports REST method/path operations that lack a
// matching statically discoverable request in the rest package.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const defaultSpecURL = "https://raw.githubusercontent.com/discord/discord-api-spec/main/specs/openapi.json"

type operation struct {
	method string
	path   string
}

func (o operation) String() string { return o.method + " " + o.path }

func normalizePath(path string) string {
	if query := strings.IndexByte(path, '?'); query >= 0 {
		path = path[:query]
	}
	path = strings.TrimRight(path, "/")
	if path == "" {
		path = "/"
	}
	for offset := 0; offset < len(path); {
		relativeStart := strings.IndexByte(path[offset:], '{')
		if relativeStart < 0 {
			break
		}
		start := offset + relativeStart
		relativeEnd := strings.IndexByte(path[start:], '}')
		if relativeEnd < 0 {
			break
		}
		end := start + relativeEnd
		if end == start+1 { // already normalized placeholder
			offset = end + 1
			continue
		}
		path = path[:start] + "{}" + path[end+1:]
		offset = start + 2
	}
	return path
}

func staticString(expr ast.Expr, variables map[string]string) (string, bool) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			return "", false
		}
		value, err := strconv.Unquote(e.Value)
		return value, err == nil
	case *ast.Ident:
		value, ok := variables[e.Name]
		return value, ok
	case *ast.ParenExpr:
		return staticString(e.X, variables)
	case *ast.BinaryExpr:
		if e.Op != token.ADD {
			return "", false
		}
		left, okLeft := staticString(e.X, variables)
		right, okRight := staticString(e.Y, variables)
		return left + right, okLeft && okRight
	case *ast.CallExpr:
		if selector, ok := e.Fun.(*ast.SelectorExpr); ok {
			switch selector.Sel.Name {
			case "String":
				return "{}", true
			case "QueryString":
				return "", true
			}
		}
		if ident, ok := e.Fun.(*ast.SelectorExpr); ok && ident.Sel.Name == "Sprintf" && len(e.Args) > 0 {
			format, ok := staticString(e.Args[0], variables)
			if !ok {
				return "", false
			}
			return strings.ReplaceAll(format, "%s", "{}"), true
		}
	}
	return "", false
}

func staticMethod(expr ast.Expr) string {
	if literal, ok := expr.(*ast.BasicLit); ok && literal.Kind == token.STRING {
		method, err := strconv.Unquote(literal.Value)
		if err == nil {
			return strings.ToUpper(method)
		}
	}
	if selector, ok := expr.(*ast.SelectorExpr); ok && strings.HasPrefix(selector.Sel.Name, "Method") {
		return strings.ToUpper(strings.TrimPrefix(selector.Sel.Name, "Method"))
	}
	return ""
}

func restOperations(restDir string) (map[operation]struct{}, error) {
	files, err := filepath.Glob(filepath.Join(restDir, "*.go"))
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	operations := make(map[operation]struct{})
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			return nil, err
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			variables := make(map[string]string)
			ast.Inspect(function.Body, func(node ast.Node) bool {
				switch statement := node.(type) {
				case *ast.AssignStmt:
					for i, lhs := range statement.Lhs {
						ident, ok := lhs.(*ast.Ident)
						if !ok || i >= len(statement.Rhs) {
							continue
						}
						rhs := statement.Rhs[i]
						if statement.Tok == token.ADD_ASSIGN {
							rhs = &ast.BinaryExpr{X: ident, Op: token.ADD, Y: rhs}
						}
						if value, ok := staticString(rhs, variables); ok {
							variables[ident.Name] = value
						}
					}
				case *ast.ValueSpec:
					for i, name := range statement.Names {
						if i < len(statement.Values) {
							if value, ok := staticString(statement.Values[i], variables); ok {
								variables[name.Name] = value
							}
						}
					}
				case *ast.CallExpr:
					selector, ok := statement.Fun.(*ast.SelectorExpr)
					if !ok || (selector.Sel.Name != "Request" && selector.Sel.Name != "RequestMultipart") || len(statement.Args) < 3 {
						return true
					}
					method := staticMethod(statement.Args[1])
					path, ok := staticString(statement.Args[2], variables)
					if method == "" || !ok || !strings.HasPrefix(path, "/") {
						return true
					}
					operations[operation{method: method, path: normalizePath(path)}] = struct{}{}
				}
				return true
			})
		}
	}
	return operations, nil
}

func exportedStructs(root string) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && path != root {
			name := entry.Name()
			if name == "dave-go" || name == "docs" || name == "node_modules" || name == "scripts" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, declaration := range file.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || !typeSpec.Name.IsExported() {
					continue
				}
				if _, ok := typeSpec.Type.(*ast.StructType); ok {
					result[typeSpec.Name.Name] = struct{}{}
				}
			}
		}
		return nil
	})
	return result, err
}

func specOperations(data []byte) (map[operation]struct{}, error) {
	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, err
	}
	operations := make(map[operation]struct{})
	for path, methods := range document.Paths {
		for method := range methods {
			upper := strings.ToUpper(method)
			if upper == "GET" || upper == "POST" || upper == "PUT" || upper == "PATCH" || upper == "DELETE" || upper == "HEAD" || upper == "OPTIONS" {
				operations[operation{method: upper, path: normalizePath(path)}] = struct{}{}
			}
		}
	}
	return operations, nil
}

func loadSpec(specPath, specURL string) ([]byte, error) {
	var data []byte
	if specPath != "" {
		var err error
		data, err = os.ReadFile(specPath)
		if err != nil {
			return nil, err
		}
	} else {
		parsed, err := url.Parse(specURL)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return nil, fmt.Errorf("invalid spec URL %q", specURL)
		}
		client := &http.Client{Timeout: 30 * time.Second}
		request, err := http.NewRequest(http.MethodGet, specURL, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("User-Agent", "discord.go-api-coverage/1")
		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("spec download returned HTTP %d", response.StatusCode)
		}
		data, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}
	var spec struct {
		Paths json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	if len(spec.Paths) == 0 {
		return nil, fmt.Errorf("spec has no paths object")
	}
	return data, nil
}

func run() error {
	specPath := flag.String("spec", "", "read an OpenAPI JSON file instead of downloading it")
	specURL := flag.String("spec-url", defaultSpecURL, "OpenAPI JSON URL")
	root := flag.String("root", ".", "repository root")
	fail := flag.Bool("fail-on-untracked", false, "exit non-zero when spec operations lack matching REST calls")
	flag.Parse()

	spec, err := loadSpec(*specPath, *specURL)
	if err != nil {
		return err
	}
	all, err := specOperations(spec)
	if err != nil {
		return err
	}
	implemented, err := restOperations(filepath.Join(*root, "rest"))
	if err != nil {
		return err
	}
	missing := make([]operation, 0)
	for operation := range all {
		if _, ok := implemented[operation]; !ok {
			missing = append(missing, operation)
		}
	}
	// Preserve stable summary ordering in CI logs.
	for i := 0; i < len(missing); i++ {
		for j := i + 1; j < len(missing); j++ {
			if missing[j].String() < missing[i].String() {
				missing[i], missing[j] = missing[j], missing[i]
			}
		}
	}
	covered := len(all) - len(missing)
	percent := 100.0
	if len(all) > 0 {
		percent = 100 * float64(covered) / float64(len(all))
	}
	var schemaDocument struct {
		Components struct {
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(spec, &schemaDocument); err != nil {
		return err
	}
	structNames, err := exportedStructs(*root)
	if err != nil {
		return err
	}
	matchedSchemas := 0
	var unmatchedSchemas []string
	for name := range schemaDocument.Components.Schemas {
		if _, ok := structNames[name]; ok {
			matchedSchemas++
		} else {
			unmatchedSchemas = append(unmatchedSchemas, name)
		}
	}
	for i := 0; i < len(unmatchedSchemas); i++ {
		for j := i + 1; j < len(unmatchedSchemas); j++ {
			if unmatchedSchemas[j] < unmatchedSchemas[i] {
				unmatchedSchemas[i], unmatchedSchemas[j] = unmatchedSchemas[j], unmatchedSchemas[i]
			}
		}
	}
	fmt.Println("# Discord REST operation inventory")
	fmt.Println()
	fmt.Printf("Spec operations: **%d**\n", len(all))
	fmt.Printf("Statically detected REST operations: **%d**\n", len(implemented))
	fmt.Printf("Exact method/path matches: **%d/%d (%.1f%%)**\n", covered, len(all), percent)
	fmt.Println()
	fmt.Printf("OpenAPI schemas: **%d**\n", len(schemaDocument.Components.Schemas))
	fmt.Printf("Exported Go structs: **%d**\n", len(structNames))
	fmt.Printf("Exact-name Go struct matches: **%d**\n", matchedSchemas)
	fmt.Println()
	fmt.Println("> Static inventory only. Exact-name model matching is a heuristic; it does not verify field parity, Gateway events, or semantic behavior. Unresolved dynamic routes may be reported as missing.")
	fmt.Println()
	fmt.Println("## Spec operations without a detected REST call")
	fmt.Println()
	for _, missingOperation := range missing {
		fmt.Printf("- `%s`\n", missingOperation.String())
	}
	if len(missing) == 0 {
		fmt.Println("- None")
	}
	fmt.Println()
	fmt.Println("## OpenAPI schemas without an exact-name exported Go struct")
	fmt.Println()
	for _, name := range unmatchedSchemas {
		fmt.Printf("- `%s`\n", name)
	}
	if len(unmatchedSchemas) == 0 {
		fmt.Println("- None")
	}
	if *fail && len(missing) > 0 {
		return fmt.Errorf("%d OpenAPI operations are not matched", len(missing))
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
