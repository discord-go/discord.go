package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	if got := normalizePath("/guilds/{guild_id}/members/{user_id}/?limit=5"); got != "/guilds/{}/members/{}" {
		t.Fatalf("normalizePath() = %q", got)
	}
}

func TestRestOperationsFindsConcatenatedRoutes(t *testing.T) {
	dir := t.TempDir()
	source := `package rest
type Client struct{}
func (c *Client) Request(any, string, string, any, any) error { return nil }
type id int
func (i id) String() string { return "42" }
func (c *Client) Create(ctx any, channelID id) {
  path := "/channels/" + channelID.String() + "/messages"
  _ = c.Request(ctx, "POST", path, nil, nil)
}`
	if err := os.WriteFile(filepath.Join(dir, "routes.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	operations, err := restOperations(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := operation{method: "POST", path: "/channels/{}/messages"}
	if _, ok := operations[want]; !ok {
		t.Fatalf("operations %v do not include %v", operations, want)
	}
}

func TestStaticStringFormatsPath(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "test.go", `package x
func f(id string) { _ = fmt.Sprintf("/guilds/%s/members", id) }`, 0)
	if err != nil {
		t.Fatal(err)
	}
	var call *ast.CallExpr
	ast.Inspect(file, func(node ast.Node) bool {
		if candidate, ok := node.(*ast.CallExpr); ok {
			call = candidate
		}
		return true
	})
	got, ok := staticString(call, nil)
	if !ok || got != "/guilds/{}/members" {
		t.Fatalf("staticString() = %q, %v", got, ok)
	}
}

func TestExportedStructs(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "models.go"), []byte("package models\ntype Guild struct{}\ntype internal struct{}\ntype GuildID string\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := exportedStructs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["Guild"]; !ok || len(got) != 1 {
		t.Fatalf("exportedStructs() = %v", got)
	}
}
