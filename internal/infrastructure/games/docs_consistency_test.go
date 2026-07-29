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

// archEndpointRe captures the game name from every `POST /<name>/exec`
// reference in docs/architecture.md. Game names are lowercase alphanumeric
// (e.g. blackjack, omahahilo, deucetoseven), matching the registry Name field.
var archEndpointRe = regexp.MustCompile(`POST /([a-z0-9]+)/exec`)

// TestArchitectureDocEndpointsMatchRegistry guards the hand-maintained Web API
// endpoint list in docs/architecture.md against the registry single source of
// truth. That list (and its spelled-out count) drifted badly — it fell 21
// games behind while still claiming "One hundred forty-eight endpoints" (see
// issue #2525) — because, unlike README.md/CLAUDE.md, no test covered it. This
// asserts a strict 1:1 mapping between registered games and the `POST
// /<name>/exec` entries in the doc, failing CI the moment one is added or
// removed without updating the doc.
func TestArchitectureDocEndpointsMatchRegistry(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "docs/architecture.md"))
	if err != nil {
		t.Fatalf("read docs/architecture.md: %v", err)
	}

	matches := archEndpointRe.FindAllSubmatch(data, -1)
	docNames := make(map[string]bool)
	for _, m := range matches {
		docNames[string(m[1])] = true
	}

	registryNames := make(map[string]bool, len(games.All()))
	for _, g := range games.All() {
		registryNames[g.Name] = true
	}

	// Catch duplicate entries: the maps above dedupe, so a game listed twice
	// would still pass the 1:1 checks below. Comparing the raw match count to
	// the (deduped) registry size surfaces a doubled endpoint.
	if len(matches) != len(registryNames) {
		t.Errorf("docs/architecture.md has %d POST /<name>/exec entries but registry has %d games — check for duplicate or stray endpoint lines", len(matches), len(registryNames))
	}

	// Every registered game must have a POST /<name>/exec entry in the doc.
	for name := range registryNames {
		if !docNames[name] {
			t.Errorf("docs/architecture.md is missing endpoint POST /%s/exec for registered game %q", name, name)
		}
	}
	// No orphan endpoints: every documented endpoint must map to a real game.
	for name := range docNames {
		if !registryNames[name] {
			t.Errorf("docs/architecture.md documents endpoint POST /%s/exec with no matching game in the registry", name)
		}
	}
}

// archEndpointCountRe captures the endpoint totals stated in
// docs/architecture.md -- the summary bullet and the table's own header line.
var archEndpointCountRe = regexp.MustCompile(`(?:\*\*Web API\*\*: |-- \*\*)(\d+)(?:\*\* in total| endpoints, one per game)`)

// TestArchitectureDocEndpointCountMatchesRegistry guards the endpoint totals
// written in prose.
//
// Until #4470 that number was spelled out in words ("Two hundred and thirty-two
// endpoints") inside a single paragraph that also held every endpoint, and
// nothing checked it: TestArchitectureDocEndpointsMatchRegistry only counts
// `POST /<name>/exec` occurrences, so the prose could say anything. Every game
// added this session updated it by hand on trust.
func TestArchitectureDocEndpointCountMatchesRegistry(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "docs/architecture.md"))
	if err != nil {
		t.Fatalf("read docs/architecture.md: %v", err)
	}

	matches := archEndpointCountRe.FindAllSubmatch(data, -1)
	if len(matches) == 0 {
		t.Fatal("docs/architecture.md states no endpoint count -- update the regex if the wording moved")
	}

	want := len(games.All())
	for _, m := range matches {
		got, err := strconv.Atoi(string(m[1]))
		if err != nil {
			t.Fatalf("unparsable endpoint count %q: %v", m[1], err)
		}
		if got != want {
			t.Errorf("docs/architecture.md states %d endpoints, registry has %d — update the doc", got, want)
		}
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
