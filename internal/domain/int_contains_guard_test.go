//go:build test

package domain_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntContainsGuard(t *testing.T) {
	const root = "."
	fset := token.NewFileSet()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read domain directory: %v", err)
	}
	files := 0
	var violations []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(root, name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		files++
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Body == nil || !isIntSliceContains(fn) {
				continue
			}
			pos := fset.Position(fn.Pos())
			violations = append(violations, fmt.Sprintf("%s:%d:%s", filepath.Base(pos.Filename), pos.Line, fn.Name.Name))
		}
	}
	if files < 300 {
		t.Fatalf("scanned only %d production files; scan may be broken", files)
	}
	if len(violations) > 0 {
		t.Errorf("manual slice containment loops found; use slices.Contains:\n  %s", strings.Join(violations, "\n  "))
	}
}

func isIntSliceContains(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 2 || fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}
	first, ok := fn.Type.Params.List[0].Type.(*ast.ArrayType)
	if !ok || first.Len != nil || !isIdent(first.Elt, "int") || !isIdent(fn.Type.Params.List[1].Type, "int") || !isIdent(fn.Type.Results.List[0].Type, "bool") {
		return false
	}
	if len(fn.Type.Params.List[0].Names) != 1 || len(fn.Type.Params.List[1].Names) != 1 {
		return false
	}
	sliceName := fn.Type.Params.List[0].Names[0].Name
	valueName := fn.Type.Params.List[1].Names[0].Name
	if len(fn.Body.List) != 2 {
		return false
	}
	rangeStmt, ok := fn.Body.List[0].(*ast.RangeStmt)
	if !ok || rangeStmt.Tok != token.DEFINE || !isIdent(rangeStmt.X, sliceName) || !isIdent(rangeStmt.Key, "_") || len(rangeStmt.Body.List) != 1 {
		return false
	}
	if rangeStmt.Value == nil {
		return false
	}
	value, ok := rangeStmt.Value.(*ast.Ident)
	if !ok {
		return false
	}
	ifStmt, ok := rangeStmt.Body.List[0].(*ast.IfStmt)
	if !ok || ifStmt.Init != nil || ifStmt.Else != nil || len(ifStmt.Body.List) != 1 {
		return false
	}
	cond, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok || cond.Op != token.EQL || !comparesPair(cond, value.Name, valueName) {
		return false
	}
	ret, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 || !isIdent(ret.Results[0], "true") {
		return false
	}
	last, ok := fn.Body.List[1].(*ast.ReturnStmt)
	return ok && len(last.Results) == 1 && isIdent(last.Results[0], "false")
}

// comparesPair reports whether cond is `a == b` or `b == a`.
func comparesPair(cond *ast.BinaryExpr, a, b string) bool {
	return isIdent(cond.X, a) && isIdent(cond.Y, b) || isIdent(cond.X, b) && isIdent(cond.Y, a)
}

func isIdent(expr ast.Expr, name string) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == name
}
