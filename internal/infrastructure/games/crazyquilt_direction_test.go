package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestCrazyQuiltDirectionMarksAgreeAcrossTheUIs guards the two ↑/↓ tables.
//
// The CUI marks an Ace-start foundation with ↑ (`crazyQuiltSeriesMark`), and the
// Web page now marks the same thing from `foundationAscending` (#5743). Both
// sides read the same flag, so a swapped pair of arrows would look consistent on
// each screen and only disagree with the other — nothing else would notice.
func TestCrazyQuiltDirectionMarksAgreeAcrossTheUIs(t *testing.T) {
	root := filepath.Join("..", "..", "..")

	cui := readFileForTest(t, filepath.Join(root, "internal", "adapter", "presenter", "CrazyQuiltCuiPresenter.go"))
	web := readFileForTest(t, filepath.Join(root, "frontend", "src", "pages", "CrazyQuiltPage.tsx"))

	// CUI: `if ascending { return "↑" }`
	cuiAscending := regexp.MustCompile(`func crazyQuiltSeriesMark\(ascending bool\) string \{\s*if ascending \{\s*return "(\\u[0-9a-fA-F]{4})"`).
		FindStringSubmatch(cui)
	if cuiAscending == nil {
		t.Fatal("crazyQuiltSeriesMark no longer returns a literal rune for the ascending case")
	}

	webAscending := regexp.MustCompile(`directionMark = ascending \? '(.)' : '(.)'`).FindStringSubmatch(web)
	if webAscending == nil {
		t.Fatal("CrazyQuiltPage.tsx no longer picks the direction mark with a literal ternary")
	}

	// CUI 側はソースに \u2191 のエスケープで書かれているので、その形で比べる。
	const cuiUpEscape = `\u2191`
	wantUp, wantDown := "↑", "↓"
	if cuiAscending[1] != cuiUpEscape {
		t.Errorf("the CUI marks an ascending foundation with %q, want %q (= %q)",
			cuiAscending[1], cuiUpEscape, wantUp)
	}
	if webAscending[1] != wantUp || webAscending[2] != wantDown {
		t.Errorf("the Web marks ascending/descending with %q/%q, want %q/%q",
			webAscending[1], webAscending[2], wantUp, wantDown)
	}
}

// readFileForTest reads a source file or fails the test.
func readFileForTest(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // test-only, fixed paths
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}
