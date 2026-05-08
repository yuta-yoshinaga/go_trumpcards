//go:build test

package presenter_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// TestNoJapaneseStringLiteralsInCuiPresenters guards the i18n migration
// (#1699 Phase 3 + Phase 4): every user-facing string in
// internal/adapter/presenter/*CuiPresenter.go must come from i18n.T/Tf
// rather than a hardcoded literal. If this test fails, route the new
// string through internal/i18n/locales/{ja,en}/<game>.json instead of
// embedding it in the Go source.
//
// The check intentionally targets only `*CuiPresenter.go` (the
// presentation layer). GoDoc comments are out of scope — only string
// literals in expression position are inspected. JA in test files,
// Web presenters, controllers, infrastructure, etc. is unaffected.
//
// Phase 3 caught eleven loader-divergence bugs precisely because the
// translation pipeline had no automated guard against drift; this test
// closes that gap so the next time someone adds a presenter the lint
// flags the bypass at PR time instead of code review.
func TestNoJapaneseStringLiteralsInCuiPresenters(t *testing.T) {
	files, err := filepath.Glob("*CuiPresenter.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no *CuiPresenter.go files found — running from wrong dir?")
	}

	type violation struct {
		file string
		pos  token.Position
		text string
	}
	var violations []violation

	fset := token.NewFileSet()
	for _, file := range files {
		af, err := parser.ParseFile(fset, file, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Errorf("parse %s: %v", file, err)
			continue
		}
		ast.Inspect(af, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			val, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			if !containsJapanese(val) {
				return true
			}
			violations = append(violations, violation{
				file: file,
				pos:  fset.Position(lit.Pos()),
				text: lit.Value,
			})
			return true
		})
	}

	if len(violations) == 0 {
		return
	}
	var b strings.Builder
	b.WriteString("hardcoded Japanese string literals in CuiPresenter source — route through i18n.T/Tf instead:\n")
	for _, v := range violations {
		b.WriteString("\n  ")
		b.WriteString(v.pos.String())
		b.WriteString(": ")
		b.WriteString(v.text)
	}
	t.Error(b.String())
}

// containsJapanese reports whether s contains hiragana, katakana, or CJK
// ideographs. Half-width katakana (U+FF65-U+FF9F) is included.
func containsJapanese(s string) bool {
	for _, r := range s {
		switch {
		case r >= 0x3041 && r <= 0x309F: // hiragana
			return true
		case r >= 0x30A0 && r <= 0x30FF: // katakana
			return true
		case r >= 0xFF65 && r <= 0xFF9F: // half-width katakana
			return true
		case unicode.Is(unicode.Han, r): // CJK ideographs
			return true
		}
	}
	return false
}
