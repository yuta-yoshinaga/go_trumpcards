//go:build test

package domain_test

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// The Cloudflare Workers are TinyGo WASM binaries, and TinyGo's reflect leaves
// Type.Implements/AssignableTo unimplemented for any interface that has at
// least one method -- it panics instead. TinyGo folds those checks away when
// the compiler can see the concrete types, which is why the Workers ran for
// months. errors.As reaches reflectlite's *dynamic* AssignableTo, and once a
// Worker binary links that in, the folding stops applying to encoding/json as
// well: every request then panicked inside Decode/Encode with a WASM trap that
// recover() cannot catch, so the Worker hung and Cloudflare answered 1101.
//
// PR #5889 made PaiGow (casino bucket) call ErrorMessageCode and took the
// casino Worker down; #5809 had already done the same to solo via
// MissMilligan. The 1101 was identical for every game in the bucket, because
// the panic is in the shared request path, not in any game's logic.
//
// TestErrorMessageCodeUnwrapsByHand pins the behaviour, and
// TestNoErrorsAsAnywhereUnderInternal pins the mechanism -- the behaviour test alone would
// stay green if someone rewrote the loop back into errors.As.

// TestErrorMessageCodeUnwrapsByHand covers the chains ErrorMessageCode has to
// walk now that it no longer delegates to errors.As.
func TestErrorMessageCodeUnwrapsByHand(t *testing.T) {
	coded := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "paigow.foulHighMustBeat", map[string]string{"n": "1"})
	phrase := domain.NewDomainError(domain.ErrInvalidPlay, "a finished sentence")

	tests := []struct {
		name       string
		err        error
		wantCode   string
		wantParams map[string]string
	}{
		{"nil error", nil, "", nil},
		{"plain error", errors.New("plain"), "", nil},
		{"coded DomainError", coded, "paigow.foulHighMustBeat", map[string]string{"n": "1"}},
		{"phrase DomainError answers empty", phrase, "", nil},
		{"coded wrapped once", fmt.Errorf("ctx: %w", coded), "paigow.foulHighMustBeat", map[string]string{"n": "1"}},
		{"coded wrapped twice", fmt.Errorf("a: %w", fmt.Errorf("b: %w", coded)), "paigow.foulHighMustBeat", map[string]string{"n": "1"}},
		{"phrase wrapped stops at the first DomainError", fmt.Errorf("ctx: %w", phrase), "", nil},
		{"sentinel only", domain.ErrInvalidPlay, "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, params := domain.ErrorMessageCode(tt.err)
			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if len(params) != len(tt.wantParams) {
				t.Fatalf("params = %v, want %v", params, tt.wantParams)
			}
			for k, v := range tt.wantParams {
				if params[k] != v {
					t.Errorf("params[%q] = %q, want %q", k, params[k], v)
				}
			}
		})
	}
}

// TestNoErrorsAsAnywhereUnderInternal fails if errors.As reappears anywhere a Worker binary
// can link. A behaviour test cannot catch that: errors.As would return exactly
// the same answers on the host and still take every Worker down.
func TestNoErrorsAsAnywhereUnderInternal(t *testing.T) {
	// ".." is internal/, walked whole rather than as a list of package names:
	// an explicit list silently stops covering a package that is added later,
	// and internal/color was already missing from one. Packages no Worker
	// links are swept up too, which is the safe direction to be wrong in --
	// errors.As has no business anywhere under internal/.
	const root = ".."

	scanned := 0
	var offenders []string
	{
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
			scanned++
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "As" {
					return true
				}
				pkg, ok := sel.X.(*ast.Ident)
				if !ok || pkg.Name != "errors" {
					return true
				}
				offenders = append(offenders, fmt.Sprintf("%s:%d", path, fset.Position(call.Pos()).Line))
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}

	// A scan that parses nothing would report success while checking nothing.
	if scanned < 100 {
		t.Fatalf("scanned only %d files; the walk root is wrong", scanned)
	}
	if len(offenders) > 0 {
		t.Errorf("errors.As is unusable in TinyGo Worker builds (it reaches reflectlite's\n"+
			"dynamic AssignableTo and makes encoding/json panic with an uncatchable WASM\n"+
			"trap -> Cloudflare 1101). Unwrap by hand with a type assertion instead.\n"+
			"Found at:\n  %s", strings.Join(offenders, "\n  "))
	}
}
