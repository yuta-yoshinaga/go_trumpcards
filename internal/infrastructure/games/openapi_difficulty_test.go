package games_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestOpenAPIDifficultyMatchesDomain prevents stale difficulty declarations
// from surviving after a game's domain configuration drops CpuDifficulty.
func TestOpenAPIDifficultyMatchesDomain(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var doc openAPIDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	controllers, err := filepath.Glob(filepath.Join(root, "internal", "adapter", "controller", "*WebController.go"))
	if err != nil {
		t.Fatal(err)
	}
	gameByRoute := map[string]string{}
	for _, p := range controllers {
		name := strings.TrimSuffix(filepath.Base(p), "WebController.go")
		gameByRoute[routeOf(name)] = name
	}
	checked := 0
	for route, op := range doc.Paths {
		m := gameExecPathRe.FindStringSubmatch(route)
		if m == nil {
			continue
		}
		game, ok := gameByRoute[m[1]]
		if !ok {
			continue
		}
		schemas := []*openAPISchema{op.Post.RequestBody.Content.JSON.Schema, op.Post.Responses.OK.Content.JSON.Schema}
		refs := map[string]bool{}
		containsDifficulty := false
		containsSetDifficulty := false
		hasCPUProperty := false
		var walk func(*openAPISchema, int)
		walk = func(s *openAPISchema, depth int) {
			if s == nil || depth > liveMaxDepth {
				return
			}
			if s.Ref != "" {
				name := strings.TrimPrefix(s.Ref, "#/components/schemas/")
				if !refs[name] {
					refs[name] = true
					walk(doc.Components.Schemas[name], depth+1)
				}
			}
			for key, child := range s.Properties {
				if key == "cpuDifficulty" || key == "difficulty" {
					containsDifficulty = true
					if key == "cpuDifficulty" {
						hasCPUProperty = true
					}
				}
				if key == "command" && child != nil {
					for _, v := range child.Enum {
						if v == "setdifficulty" {
							containsSetDifficulty = true
						}
					}
				}
				walk(child, depth+1)
			}
			walk(s.Items, depth+1)
			for _, group := range [][]*openAPISchema{s.AllOf, s.OneOf, s.AnyOf} {
				for _, child := range group {
					walk(child, depth+1)
				}
			}
		}
		for _, s := range schemas {
			walk(s, 0)
		}
		if !hasCPUProperty && (!containsDifficulty || !containsSetDifficulty) {
			continue
		}
		checked++
		path := filepath.Join(root, "internal", "domain", game+"Config.go")
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if os.IsNotExist(err) {
			path = filepath.Join(root, "internal", "domain", game+".go")
			f, err = parser.ParseFile(token.NewFileSet(), path, nil, 0)
		}
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		found := false
		var alias string
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok || ts.Name.Name != game+"Config" {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				if id, ok := ts.Type.(*ast.Ident); ok {
					alias = id.Name
				}
				return false
			}
			for _, field := range st.Fields.List {
				for _, name := range field.Names {
					if name.Name == "CpuDifficulty" {
						found = true
					}
				}
			}
			return false
		})
		if alias != "" {
			basePath := filepath.Join(root, "internal", "domain", alias+".go")
			base, parseErr := parser.ParseFile(token.NewFileSet(), basePath, nil, 0)
			if parseErr == nil {
				ast.Inspect(base, func(n ast.Node) bool {
					ts, ok := n.(*ast.TypeSpec)
					if !ok || ts.Name.Name != alias {
						return true
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						return false
					}
					for _, field := range st.Fields.List {
						for _, name := range field.Names {
							if name.Name == "CpuDifficulty" {
								found = true
							}
						}
					}
					return false
				})
			}
		}
		if !found {
			t.Errorf("/%s/exec declares CPU difficulty but %s has no CpuDifficulty field", m[1], game+"Config")
		}
	}
	if checked == 0 {
		t.Fatal("no OpenAPI difficulty declarations were checked")
	}
}
