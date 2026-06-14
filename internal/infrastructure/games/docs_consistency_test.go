package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
)

// repoRoot is the module root relative to this test's working directory
// (the package dir internal/infrastructure/games).
const repoRoot = "../../.."

// readDocNumber reads the file at path (relative to repoRoot), applies re —
// which must capture exactly one numeric group — and returns the captured
// integer. A missing match fails loudly so a reworded sentence surfaces here
// (update the regex) rather than silently skipping the assertion.
func readDocNumber(t *testing.T, path string, re *regexp.Regexp) int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot, path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	m := re.FindSubmatch(data)
	if m == nil {
		t.Fatalf("%s: pattern %q not found — did the wording change? update the regex in docs_consistency_test.go", path, re)
	}
	n, err := strconv.Atoi(string(m[1]))
	if err != nil {
		t.Fatalf("%s: matched %q is not an integer: %v", path, m[1], err)
	}
	return n
}

// TestDocGameCountsMatchRegistry guards the hand-maintained game-count numbers
// in README.md and CLAUDE.md against the registry single source of truth.
// These prose counts drift on every new game (see issue #2474, where README
// still said 126 after the registry reached 132); this test fails CI the
// moment they fall out of sync instead of waiting for a manual drift sweep.
func TestDocGameCountsMatchRegistry(t *testing.T) {
	total := len(games.All())

	cases := []struct {
		name string
		path string
		re   *regexp.Regexp
	}{
		{"README.md Features prose", "README.md", regexp.MustCompile(`実装した(\d+)種類`)},
		{"CLAUDE.md intro line", "CLAUDE.md", regexp.MustCompile(`Go implementations of (\d+) trump card game`)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := readDocNumber(t, c.path, c.re); got != total {
				t.Errorf("%s says %d games, registry has %d — update the doc (or the regex if the wording moved)", c.name, got, total)
			}
		})
	}
}

// TestPerGameManualsMatchRegistry asserts a strict 1:1 mapping between
// registered games and the per-game manuals under docs/manual/{cui,web}. It
// catches both a missing manual for a freshly added game and an orphan manual
// left behind by a rename or removal. Template files live at
// docs/manual/{cui,web}_template.md (outside these dirs) and are not counted.
func TestPerGameManualsMatchRegistry(t *testing.T) {
	all := games.All()
	names := make(map[string]bool, len(all))
	for _, g := range all {
		names[g.Name] = true
	}

	for _, dir := range []string{"docs/manual/cui", "docs/manual/web"} {
		t.Run(dir, func(t *testing.T) {
			// Every registered game must have a manual in this dir.
			for _, g := range all {
				p := filepath.Join(repoRoot, dir, g.Name+".md")
				if _, err := os.Stat(p); err != nil {
					t.Errorf("missing manual %s/%s.md for registered game %q", dir, g.Name, g.Name)
				}
			}
			// No orphan manuals: every <name>.md must map to a registered game.
			entries, err := os.ReadDir(filepath.Join(repoRoot, dir))
			if err != nil {
				t.Fatalf("read dir %s: %v", dir, err)
			}
			for _, e := range entries {
				name, ok := strings.CutSuffix(e.Name(), ".md")
				if !ok {
					t.Errorf("%s/%s is not a .md manual file", dir, e.Name())
					continue
				}
				if !names[name] {
					t.Errorf("orphan manual %s/%s.md has no matching game in the registry", dir, name)
				}
			}
		})
	}
}
