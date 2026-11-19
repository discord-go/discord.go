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

func TestRestOperationsFindsEscapedReaction(t *testing.T) {
	dir := t.TempDir()
	source := `package rest
func f(ctx any, channelID ID, emoji string) {
 path := "/channels/" + channelID.String() + "/reactions/" + url.PathEscape(emoji) + params.query()
 c.Request(ctx, "GET", path, nil, nil)
}`
	if err := os.WriteFile(filepath.Join(dir, "routes.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	operations, err := restOperations(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := operation{method: "GET", path: "/channels/{}/reactions/{}"}
	if _, ok := operations[want]; !ok {
		t.Fatalf("operations %v do not include %v", operations, want)
	}
}

func TestRestOperationsFindsLocalItoaRoute(t *testing.T) {
	dir := t.TempDir()
	source := `package rest
func f(ctx any, channelID ID, messageID ID, answerID int, params Params) {
 path := "/channels/" + channelID.String() + "/polls/" + messageID.String() + "/answers/" + itoa(answerID) + params.QueryString()
 c.Request(ctx, "GET", path, nil, nil)
}`
	if err := os.WriteFile(filepath.Join(dir, "routes.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	operations, err := restOperations(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := operation{method: "GET", path: "/channels/{}/polls/{}/answers/{}"}
	if _, ok := operations[want]; !ok {
		t.Fatalf("operations %v do not include %v", operations, want)
	}
}

func TestRestOperationsFindsStringParameterRoute(t *testing.T) {
	dir := t.TempDir()
	source := `package rest
func f(ctx any, code string) {
 path := "/invites/" + code
 c.Request(ctx, "DELETE", path, nil, nil)
}`
	if err := os.WriteFile(filepath.Join(dir, "routes.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	operations, err := restOperations(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := operation{method: "DELETE", path: "/invites/{}"}
	if _, ok := operations[want]; !ok {
		t.Fatalf("operations %v do not include %v", operations, want)
	}
}

func TestRestOperationsFindsNoAuthRoute(t *testing.T) {
	dir := t.TempDir()
	source := `package rest
func f(ctx any, webhookID ID, token string) {
 path := "/webhooks/" + webhookID.String() + "/" + token
 c.RequestNoAuth(ctx, "GET", path, nil, nil)
}`
	if err := os.WriteFile(filepath.Join(dir, "routes.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	operations, err := restOperations(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := operation{method: "GET", path: "/webhooks/{}/{}"}
	if _, ok := operations[want]; !ok {
		t.Fatalf("operations %v do not include %v", operations, want)
	}
}

func TestRestOperationsFindsRepresentativeRoutes(t *testing.T) {
	operations, err := restOperations(filepath.Join("..", "..", "rest"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []operation{
		{method: "GET", path: "/channels/{}/messages"},
		{method: "POST", path: "/channels/{}/messages"},
		{method: "GET", path: "/channels/{}/polls/{}/answers/{}"},
		{method: "DELETE", path: "/invites/{}"},
		{method: "GET", path: "/webhooks/{}/{}"},
		{method: "POST", path: "/interactions/{}/{}/callback"},
		{method: "GET", path: "/users/@me"},
		{method: "GET", path: "/gateway/bot"},
	} {
		if _, ok := operations[want]; !ok {
			t.Errorf("representative route %v not detected", want)
		}
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
