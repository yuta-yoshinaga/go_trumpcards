package games_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// docs/design/backend.md is ~90 Mermaid diagrams describing the Go backend, and
// until now nothing checked that the names in them still exist. The 2026-08-08
// sweep (#5170) found it naming two deleted types, a renamed SessionStore
// method, a renamed controller entry point, and roughly 150 methods that had
// been renamed out from under it. TestDesignDocCountsMatchRegistry (#5173)
// covers the totals; this covers the identifiers, which is where the real rot
// was.
//
// The check is structural, not textual: the Go side is parsed with go/ast so
// that interfaces, generics, embedded types and constructors are all modelled
// properly. Approximating any of those with a grep produces false positives,
// and a guard that cries wolf gets deleted.

// designDocPath is the diagram-heavy backend design document under test.
const designDocPath = "docs/design/backend.md"

// goSourceRoot is the tree whose exported surface the document describes.
const goSourceRoot = "internal"

// placeholderClasses are diagram nodes that stand for "the per-game
// implementation" rather than naming one concrete Go type. They are a real
// documentation device here -- there are 264 games and the diagrams show the
// shape once -- so they are expected to have no counterpart in code.
var placeholderClasses = map[string]bool{
	"GameCui":           true,
	"GameCuiController": true,
	"GameCuiPresenter":  true,
	"GameInteractor":    true,
	"GameInteractorIF":  true,
	"GameWebPresenter":  true,
}

var (
	// mermaidBlockRe captures the body of a fenced mermaid block.
	mermaidBlockRe = regexp.MustCompile("(?s)```mermaid\n(.*?)```")
	// classDiagramRe detects the diagram kind, since only class diagrams
	// declare types and members.
	classDiagramRe = regexp.MustCompile(`(?m)^\s*classDiagram`)
	// classOpenRe matches `class Foo {`, classBareRe a member-less `class Foo`.
	classOpenRe = regexp.MustCompile(`^\s*class\s+([A-Za-z0-9_]+)\s*\{`)
	classBareRe = regexp.MustCompile(`^\s*class\s+([A-Za-z0-9_]+)\s*$`)
	// memberFuncRe matches a method entry such as `+GetPhase() int`. Fields
	// (no parentheses) are deliberately not checked: the diagrams routinely
	// rename a field for readability, and unexported fields are an
	// implementation detail the document is allowed to paraphrase.
	memberFuncRe = regexp.MustCompile(`^\s*[+\-#~]([A-Za-z0-9_]+)\s*\(`)
	// closeBraceRe ends a class body.
	closeBraceRe = regexp.MustCompile(`^\s*\}`)
)

// parseDesignDocClasses returns the classes declared in the document's class
// diagrams, mapped to the method names listed under each.
func parseDesignDocClasses(t *testing.T, path string) map[string]map[string]bool {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot, path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	classes := map[string]map[string]bool{}
	for _, block := range mermaidBlockRe.FindAllSubmatch(data, -1) {
		body := string(block[1])
		if !classDiagramRe.MatchString(body) {
			continue
		}
		current := ""
		for line := range strings.SplitSeq(body, "\n") {
			if m := classOpenRe.FindStringSubmatch(line); m != nil {
				current = m[1]
				if classes[current] == nil {
					classes[current] = map[string]bool{}
				}
				continue
			}
			if current != "" && closeBraceRe.MatchString(line) {
				current = ""
				continue
			}
			if m := classBareRe.FindStringSubmatch(line); m != nil {
				if classes[m[1]] == nil {
					classes[m[1]] = map[string]bool{}
				}
				continue
			}
			if current != "" {
				if m := memberFuncRe.FindStringSubmatch(line); m != nil {
					classes[current][m[1]] = true
				}
			}
		}
	}
	return classes
}

// goSurface is the parsed Go type surface: which named types exist, which
// methods each has, and which types each struct embeds.
type goSurface struct {
	types    map[string]bool
	methods  map[string]map[string]bool
	embedded map[string][]string
}

// has reports whether typ (or anything it embeds) provides method. Embedding is
// resolved transitively because the diagrams legitimately list a promoted
// method -- GamePlayer's accessors show up on the per-game player classes -- and
// treating those as missing would be a false positive on correct docs.
func (s *goSurface) has(typ, method string, seen map[string]bool) bool {
	if seen[typ] {
		return false
	}
	seen[typ] = true
	if s.methods[typ][method] {
		return true
	}
	for _, base := range s.embedded[typ] {
		if s.has(base, method, seen) {
			return true
		}
	}
	return false
}

// receiverName extracts the bare type name from a method receiver, unwrapping
// pointers and generic instantiations (`*Foo[T]` -> `Foo`).
func receiverName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return receiverName(e.X)
	case *ast.IndexExpr:
		return receiverName(e.X)
	case *ast.IndexListExpr:
		return receiverName(e.X)
	case *ast.Ident:
		return e.Name
	}
	return ""
}

// embeddedName extracts the type name of an embedded field, unwrapping
// pointers, generics and qualified names.
func embeddedName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return embeddedName(e.X)
	case *ast.IndexExpr:
		return embeddedName(e.X)
	case *ast.IndexListExpr:
		return embeddedName(e.X)
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.Ident:
		return e.Name
	}
	return ""
}

// parseGoSurface walks root and records every named type, its methods, its
// embedded types, and its constructors. Test files are skipped: the document
// describes production code.
func parseGoSurface(t *testing.T, root string) *goSurface {
	t.Helper()
	s := &goSurface{
		types:    map[string]bool{},
		methods:  map[string]map[string]bool{},
		embedded: map[string][]string{},
	}
	addMethod := func(typ, name string) {
		if s.methods[typ] == nil {
			s.methods[typ] = map[string]bool{}
		}
		s.methods[typ][name] = true
	}

	fset := token.NewFileSet()
	walkErr := filepath.Walk(filepath.Join(repoRoot, root), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		// Build tags split this tree (server vs WASM); parse every file
		// regardless so the surface is the union, not one build's view.
		file, perr := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if perr != nil {
			return nil // an unparseable file is not this guard's business
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					s.types[ts.Name.Name] = true
					switch typ := ts.Type.(type) {
					case *ast.InterfaceType:
						for _, m := range typ.Methods.List {
							for _, n := range m.Names {
								addMethod(ts.Name.Name, n.Name)
							}
							if len(m.Names) == 0 {
								if base := embeddedName(m.Type); base != "" {
									s.embedded[ts.Name.Name] = append(s.embedded[ts.Name.Name], base)
								}
							}
						}
					case *ast.StructType:
						for _, f := range typ.Fields.List {
							if len(f.Names) == 0 {
								if base := embeddedName(f.Type); base != "" {
									s.embedded[ts.Name.Name] = append(s.embedded[ts.Name.Name], base)
								}
							}
						}
					}
				}
			case *ast.FuncDecl:
				if d.Recv != nil && len(d.Recv.List) == 1 {
					if typ := receiverName(d.Recv.List[0].Type); typ != "" {
						addMethod(typ, d.Name.Name)
					}
					continue
				}
				// A package-level `NewFoo` is shown inside class Foo in the
				// diagrams, which is the conventional way to draw a
				// constructor. Credit it to Foo so that is not a false hit.
				if typ, ok := strings.CutPrefix(d.Name.Name, "New"); ok {
					addMethod(typ, d.Name.Name)
				}
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk %s: %v", root, walkErr)
	}
	return s
}

// TestDesignDocClassesExistInGo asserts that every class named in a
// docs/design/backend.md class diagram is a real Go type.
//
// This is what would have caught EuchreTrickCard and BridgeTrickCard, which
// stayed in the diagrams for months after both were replaced by the shared
// TrickCard -- while a comment in the same file already said the migration had
// happened.
func TestDesignDocClassesExistInGo(t *testing.T) {
	classes := parseDesignDocClasses(t, designDocPath)
	if len(classes) < 50 {
		t.Fatalf("only %d classes parsed from %s -- the diagram format changed; update the regexes", len(classes), designDocPath)
	}
	surface := parseGoSurface(t, goSourceRoot)
	if len(surface.types) < 500 {
		t.Fatalf("only %d Go types parsed from %s/ -- the walk is broken", len(surface.types), goSourceRoot)
	}

	var unknown []string
	for name := range classes {
		if placeholderClasses[name] || surface.types[name] {
			continue
		}
		unknown = append(unknown, name)
	}
	sort.Strings(unknown)

	if len(unknown) > 0 {
		t.Errorf("classes in %s with no Go type: %v\n"+
			"Either the type was renamed or deleted (update the diagram), or it is a\n"+
			"deliberate per-game placeholder (add it to placeholderClasses).", designDocPath, unknown)
	}
}

// TestDesignDocMethodsExistInGo asserts that every method listed under a class
// that does map to a real Go type actually exists on it.
//
// This is the check that bites: the sweep found SessionStore.GetOrCreate,
// GameWebController.Handle, GameManager.Switch and a long tail of renames
// (Phase -> GetPhase, ActionLog -> GetActionLog) still documented years after
// the code moved on.
func TestDesignDocMethodsExistInGo(t *testing.T) {
	classes := parseDesignDocClasses(t, designDocPath)
	surface := parseGoSurface(t, goSourceRoot)

	checked := 0
	var missing []string
	for class, members := range classes {
		if !surface.types[class] {
			continue // covered by TestDesignDocClassesExistInGo
		}
		for method := range members {
			checked++
			if !surface.has(class, method, map[string]bool{}) {
				missing = append(missing, class+"."+method)
			}
		}
	}
	if checked < 200 {
		t.Fatalf("only %d methods checked -- the member regex stopped matching; update memberFuncRe", checked)
	}
	sort.Strings(missing)

	if len(missing) > 0 {
		t.Errorf("methods in %s that do not exist on the Go type (%d of %d checked):\n  %s\n"+
			"Rename them in the diagram to match the code.", designDocPath, len(missing), checked, strings.Join(missing, "\n  "))
	}
}
