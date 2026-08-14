package games_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
)

// The guards in this file close the holes that the 2026-08-08 drift sweep
// (issues #5170/#5172) found. Every discrepancy that sweep turned up sat in a
// doc nothing asserted on; every doc that had a guard was clean. The lesson is
// not "be more careful" -- it is that an unguarded list drifts, so each of
// these asserts a *set* or a *count* against the registry rather than trusting
// a reviewer to notice a missing row.
//
// Each guard fails loudly when it parses nothing. A scan that silently matches
// zero rows reports success while checking nothing, which is the failure mode
// that let README fall 30 games behind while its own prose count stayed green.

// registryNames returns the registry game names as a set.
func registryNames() map[string]bool {
	set := map[string]bool{}
	for _, g := range games.All() {
		set[g.Name] = true
	}
	return set
}

// diffSets returns the elements of want missing from got, and the elements of
// got that want does not contain, both sorted for a stable failure message.
func diffSets(want, got map[string]bool) (missing, extra []string) {
	for name := range want {
		if !got[name] {
			missing = append(missing, name)
		}
	}
	for name := range got {
		if !want[name] {
			extra = append(extra, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}

// readmeGameRowRe captures the command column of a row in the README Features
// table, e.g. "| ブラックジャック (BlackJack) | `blackjack` | [CUI](...) / [Web](...) |".
var readmeGameRowRe = regexp.MustCompile("(?m)^\\| .+ \\| `([a-z0-9]+)` \\| \\[CUI\\]")

// TestReadmeGameTableMatchesRegistry asserts that the README Features table
// lists exactly the registered games.
//
// TestDocGameCountsMatchRegistry already guards the prose "実装したN種類"
// immediately above this table -- and that is precisely why the table drifted
// undetected. A guard on the number is not a guard on the list: the prose said
// 264 while the table held 234 rows, so 30 new games shipped without a row and
// CI stayed green every time (issue #5170). Assert the set, not the count.
func TestReadmeGameTableMatchesRegistry(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}

	listed := map[string]bool{}
	for _, m := range readmeGameRowRe.FindAllSubmatch(data, -1) {
		listed[string(m[1])] = true
	}
	if len(listed) == 0 {
		t.Fatal("no game rows parsed from the README Features table -- the table format changed; update readmeGameRowRe")
	}

	missing, extra := diffSets(registryNames(), listed)
	if len(missing) > 0 {
		t.Errorf("registered games with no README table row: %v -- add a row to the Features table in README.md", missing)
	}
	if len(extra) > 0 {
		t.Errorf("README table rows for games that are not registered: %v -- a rename or removal left them behind", extra)
	}
}

// gamesDocBulletRe matches one game entry in docs/games.md ("- **Name**: ...").
var gamesDocBulletRe = regexp.MustCompile(`(?m)^- \*\*`)

// TestGamesDocBulletCountMatchesRegistry asserts that docs/games.md holds one
// bullet per registered game. It had 255 for 264 games before issue #5170.
//
// This checks the count rather than the set, deliberately. The bullets are
// keyed by display name, not registry key, and those names carry diacritics
// and punctuation the key drops (Chinchón/chinchon, Pig's Tail/pigtail,
// Königrufen/koenigrufen), so a naive key match reports false gaps on entries
// that are present. A count catches the observed failure -- games added without
// a bullet -- and does so without a name-normalisation table that would itself
// drift. A same-size swap slips through; that has not happened, and a wrong
// guard is worse than a narrow one.
func TestGamesDocBulletCountMatchesRegistry(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "docs/games.md"))
	if err != nil {
		t.Fatalf("read docs/games.md: %v", err)
	}

	got := len(gamesDocBulletRe.FindAll(data, -1))
	if got == 0 {
		t.Fatal("no `- **Name**:` bullets parsed from docs/games.md -- the format changed; update gamesDocBulletRe")
	}
	if want := len(games.All()); got != want {
		t.Errorf("docs/games.md has %d game bullets, registry has %d -- every game needs an entry", got, want)
	}
}

// openapiTagRe captures a tag declared in the root `tags:` block of
// api/openapi.yaml. The two-space indent distinguishes these declarations from
// the eight-space per-operation `tags:` references.
var openapiTagRe = regexp.MustCompile(`(?m)^  - name: ([a-z0-9]+)\r?$`)

// TestOpenAPITagsMatchRegistry asserts that the root `tags:` block declares
// exactly one tag per registered game.
//
// TestOpenAPIMatchesRegistry guards the `paths:` block, and that block was
// perfect; the tag block sat at 259 of 264 (aluette, ganjifa, minchiate,
// tarocchini, vira) because nothing looked at it. Each operation references its
// game as a tag, so an undeclared tag renders in Swagger UI with no
// description.
//
// api/openapi.yaml is CRLF, hence the trailing \r? -- without it an anchored
// pattern matches nothing and this guard would pass while checking zero tags.
func TestOpenAPITagsMatchRegistry(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "api/openapi.yaml"))
	if err != nil {
		t.Fatalf("read api/openapi.yaml: %v", err)
	}

	declared := map[string]bool{}
	for _, m := range openapiTagRe.FindAllSubmatch(data, -1) {
		declared[string(m[1])] = true
	}
	if len(declared) == 0 {
		t.Fatal("no tags parsed from the root tags: block of api/openapi.yaml -- the format or line endings changed; update openapiTagRe")
	}

	missing, extra := diffSets(registryNames(), declared)
	if len(missing) > 0 {
		t.Errorf("registered games with no OpenAPI tag declaration: %v -- add `- name: <game>` to the root tags: block", missing)
	}
	if len(extra) > 0 {
		t.Errorf("OpenAPI tag declarations for games that are not registered: %v -- a rename or removal left them behind", extra)
	}
}

// TestDesignDocCountsMatchRegistry guards the hand-written totals embedded in
// the Mermaid diagrams and notes of docs/design/{backend,frontend}.md.
//
// Those two files are the worst drift site in the repo: nothing asserts on
// them, and the sweep in issue #5170 found them still saying 219 games long
// after the registry reached 264, alongside deleted types and renamed methods.
// Identifier-level checking is a larger job (tracked in #5172); the counts are
// cheap, and they are what goes stale on literally every new game.
//
// The patterns are scoped rather than a blanket `(\d+)ゲーム` scan: the docs
// legitimately contain phrases like "2 ゲーム" and "3 ゲーム" that are not
// totals, so a blanket match would fail on correct docs.
func TestDesignDocCountsMatchRegistry(t *testing.T) {
	total := len(games.All())
	// Per game there is one CUI and one Web implementation.
	perGamePair := total * 2
	// One i18n namespace per game plus common, tutorial and discover.
	namespaces := total + 3

	cases := []struct {
		name string
		path string
		re   *regexp.Regexp
		want []int
	}{
		{"backend.md 全Nゲーム", "docs/design/backend.md", regexp.MustCompile(`全(\d+)ゲーム`), []int{total}},
		{"backend.md controller/presenter pairs", "docs/design/backend.md", regexp.MustCompile(`(\d+)ゲーム × CUI/Web = (\d+)`), []int{total, perGamePair}},
		{"backend.md holds N", "docs/design/backend.md", regexp.MustCompile(`holds (\d+) (?:controllers|games)`), []int{total}},
		{"frontend.md 全Nゲーム", "docs/design/frontend.md", regexp.MustCompile(`全(\d+)ゲーム`), []int{total}},
		{"frontend.md +Routes", "docs/design/frontend.md", regexp.MustCompile(`\+Routes \((\d+)ゲーム\)`), []int{total}},
		{"frontend.md routes to N game pages", "docs/design/frontend.md", regexp.MustCompile(`routes to (\d+) game pages`), []int{total}},
		{"frontend.md i18n namespaces", "docs/design/frontend.md", regexp.MustCompile(`(\d+)名前空間: common \+ (\d+)ゲーム固有`), []int{namespaces, total}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(repoRoot, c.path))
			if err != nil {
				t.Fatalf("read %s: %v", c.path, err)
			}
			matches := c.re.FindAllStringSubmatch(string(data), -1)
			if len(matches) == 0 {
				t.Fatalf("%s: pattern %q matched nothing -- the wording changed; update the regex in docs_drift_guards_test.go", c.path, c.re)
			}
			for _, m := range matches {
				for i, want := range c.want {
					got := m[i+1]
					if got != fmt.Sprint(want) {
						t.Errorf("%s: %q says %s where the registry implies %d", c.path, m[0], got, want)
					}
				}
			}
		})
	}
}
