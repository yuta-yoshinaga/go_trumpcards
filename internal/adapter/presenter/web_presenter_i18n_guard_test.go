package presenter

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const webPresenterI18nWorkerMessage = "Worker では i18n が空のスタブなので、Web 応答にはコードと値を送りフロントで訳す (#11901)"

type webPresenterI18nCall struct {
	file string
	line int
	name string
	arg  string
}

// Only these calls are safe because the frontend translates their code fields:
// Mao's ruleHintCode is translated by maoRuleHint.ts.
// Tichu's messageCode is translated by the frontend; Message is only a fallback.
var allowedWebPresenterI18nCalls = map[string]string{
	"MaoWebPresenter.go|\"mao.\" + resObj.RuleHintCode": "",
	"TichuWebPresenter.go|resObj.MessageCode":           "Tf",
}

func collectWebPresenterI18nCalls(filename, source string) ([]webPresenterI18nCall, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, source, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	var calls []webPresenterI18nCall
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "T" && selector.Sel.Name != "Tf" {
			return true
		}
		pkg, ok := selector.X.(*ast.Ident)
		if !ok || pkg.Name != "i18n" {
			return true
		}
		var arg bytes.Buffer
		if err := format.Node(&arg, fset, call.Args[0]); err != nil {
			return true
		}
		calls = append(calls, webPresenterI18nCall{
			file: filepath.Base(filename), line: fset.Position(call.Pos()).Line,
			name: selector.Sel.Name, arg: arg.String(),
		})
		return true
	})
	return calls, nil
}

func validateWebPresenterI18nCalls(filename string, calls []webPresenterI18nCall, seen map[string]int) []string {
	var problems []string
	for _, call := range calls {
		key := call.file + "|" + call.arg
		allowedName, allowed := allowedWebPresenterI18nCalls[key]
		if allowed && allowedName != "" && allowedName != call.name {
			allowed = false
		}
		if allowed {
			seen[key]++
			continue
		}
		problems = append(problems, fmt.Sprintf("%s:%d: i18n.%s(%s): %s", filename, call.line, call.name, call.arg, webPresenterI18nWorkerMessage))
	}
	return problems
}

func TestWebPresenterI18nGuard(t *testing.T) {
	paths, err := filepath.Glob("*WebPresenter.go")
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]int, len(allowedWebPresenterI18nCalls))
	var problems []string
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		calls, err := collectWebPresenterI18nCalls(path, string(source))
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		problems = append(problems, validateWebPresenterI18nCalls(path, calls, seen)...)
	}
	if len(paths) < 200 {
		t.Errorf("scanned only %d *WebPresenter.go files; expected at least 200", len(paths))
	}
	for key := range allowedWebPresenterI18nCalls {
		if seen[key] == 0 {
			problems = append(problems, fmt.Sprintf("allowed call %q was not found; remove it from the allowlist", key))
		}
	}
	if len(problems) > 0 {
		t.Fatalf("unexpected Web presenter i18n calls:\n%s", strings.Join(problems, "\n"))
	}
}

func TestWebPresenterI18nGuardDetectsUnallowedCall(t *testing.T) {
	source := `package presenter
func example() { _ = i18n.T("unapproved.key") }
`
	calls, err := collectWebPresenterI18nCalls("ExampleWebPresenter.go", source)
	if err != nil {
		t.Fatal(err)
	}
	problems := validateWebPresenterI18nCalls("ExampleWebPresenter.go", calls, map[string]int{})
	if len(problems) != 1 || !strings.Contains(problems[0], "ExampleWebPresenter.go:2") || !strings.Contains(problems[0], webPresenterI18nWorkerMessage) {
		t.Fatalf("expected unallowed i18n call diagnostic, got %v", problems)
	}
}
