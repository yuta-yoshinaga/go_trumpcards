package games_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
)

// TestWorkerRegistrationsCoverAllGames asserts that every game in
// games.ByCategory(c) has a matching games.RegisterKVGame call in the
// corresponding worker sub-package (casino, classic, solo, extra) — and conversely,
// that no sub-package registers a name absent from the registry.
//
// The worker init functions live under `//go:build js && wasm`, so a normal
// `go test ./...` run cannot execute them and registry.RegisterCategory's
// runtime nil-check only fires at `wrangler deploy` time. ADR-0031 (option 3)
// fills that gap with AST-level source inspection: zero new CI infrastructure,
// zero changes to the existing build-tag layout. See docs/adr/0031-registry-consolidation.md.
//
// We scan the whole sub-package directory (not just <cat>.go) so the guard
// keeps working if registration logic is later split across multiple files.
func TestWorkerRegistrationsCoverAllGames(t *testing.T) {
	cases := []struct {
		cat games.Category
		dir string
	}{
		{games.CategoryCasino, "casino"},
		{games.CategoryClassic, "classic"},
		{games.CategorySolo, "solo"},
	}

	for _, c := range cases {
		t.Run(c.cat.String(), func(t *testing.T) {
			expected := registryNamesForCategory(c.cat)
			registered := parseRegisterKVGameNames(t, c.dir)
			assertEqualNameSets(t, c.cat.String(), expected, registered)
		})
	}
}

func registryNamesForCategory(cat games.Category) []string {
	entries := games.ByCategory(cat)
	out := make([]string, 0, len(entries))
	for _, g := range entries {
		out = append(out, g.Name)
	}
	sort.Strings(out)
	return out
}

// parseRegisterKVGameNames returns the first string-literal argument of every
// RegisterKVGame call across all .go files in dir, sorted. Tolerates both
// qualified (games.RegisterKVGame) and unqualified (RegisterKVGame) call
// forms so it keeps working if someone reorganises imports in the
// sub-package, and it walks every file in the directory so the guard
// survives a future split across multiple files (e.g. casino_gen.go).
//
// We deliberately use filepath.Glob + parser.ParseFile rather than
// parser.ParseDir (deprecated) or go/packages (heavier dependency); the
// `//go:build js && wasm` tag on the source files does not matter because
// we only need the AST, never compile it.
func parseRegisterKVGameNames(t *testing.T, dir string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("glob %s: %v", dir, err)
	}
	if len(matches) == 0 {
		t.Fatalf("no .go files found in %s", dir)
	}

	fset := token.NewFileSet()
	var names []string
	for _, path := range matches {
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			if !isRegisterKVGameCall(call.Fun) {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				t.Errorf("%s: RegisterKVGame first argument is not a string literal (at %s)",
					path, fset.Position(call.Pos()))
				return true
			}
			s, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Errorf("%s: cannot unquote %s at %s: %v",
					path, lit.Value, fset.Position(lit.Pos()), err)
				return true
			}
			names = append(names, s)
			return true
		})
	}

	sort.Strings(names)
	return names
}

func isRegisterKVGameCall(fun ast.Expr) bool {
	switch f := fun.(type) {
	case *ast.SelectorExpr:
		return f.Sel.Name == "RegisterKVGame"
	case *ast.Ident:
		return f.Name == "RegisterKVGame"
	default:
		return false
	}
}

// assertEqualNameSets reports asymmetric differences with messages that point
// the reader at the exact action needed (add a RegisterKVGame call, or remove
// a stray one) — the failure has to be self-explanatory because it might be
// the first time someone touches this layer.
func assertEqualNameSets(t *testing.T, category string, want, got []string) {
	t.Helper()
	wantSet := toStringSet(want)
	gotSet := toStringSet(got)

	for name := range wantSet {
		if !gotSet[name] {
			t.Errorf("%s worker is missing games.RegisterKVGame(%q, ...) — declared in registry.go but not wired in the %[1]s sub-package",
				category, name)
		}
	}
	for name := range gotSet {
		if !wantSet[name] {
			t.Errorf("%s worker registers %q but the registry has no matching entry in this category — fix Category in registry.go or remove the stray RegisterKVGame call",
				category, name)
		}
	}
}

func toStringSet(s []string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, x := range s {
		m[x] = true
	}
	return m
}
