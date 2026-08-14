package games_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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

// implTypeRe matches a per-game implementation type declaration such as
// `type BlackJackCuiController struct`.
func implTypeRe(suffix string) *regexp.Regexp {
	return regexp.MustCompile(`(?m)^type ([A-Za-z0-9_]+` + suffix + `) `)
}

// countImplTypes returns how many distinct types under dir end in suffix.
// Production files only: the document describes shipped code.
func countImplTypes(t *testing.T, dir, suffix string) int {
	t.Helper()
	re := implTypeRe(suffix)
	seen := map[string]bool{}
	root := filepath.Join(repoRoot, dir)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		for _, m := range re.FindAllSubmatch(data, -1) {
			seen[string(m[1])] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	if len(seen) < 100 {
		t.Fatalf("only %d %s types found under %s -- the declaration regex stopped matching", len(seen), suffix, dir)
	}
	return len(seen)
}

// TestDesignDocImplementationCountsMatchCode asserts the *type* counts the
// design document states for the per-game controllers and presenters.
//
// TestDesignDocCountsMatchRegistry already guards `318ゲーム × CUI/Web = 636`,
// but that is a count of *bindings*, not of types: several controllers serve
// more than one game (`NewOmahaWebController` is bound for 5, VideoPoker /
// SevenCardStud / Pineapple for 3 each, BlackJack / FreeCell / FiveCardStud for
// 2), so there are 611 controller and 612 presenter types, not 636 of each.
//
// Reading 636 as "636 implementations" is what #5350 flagged and could not fix:
// correcting the prose alone would have contradicted the binding guard, so the
// number needed a guard of its own before the wording could say both things.
func TestDesignDocImplementationCountsMatchCode(t *testing.T) {
	cui := countImplTypes(t, "internal/adapter/controller", "CuiController")
	web := countImplTypes(t, "internal/adapter/controller", "WebController")
	cuiP := countImplTypes(t, "internal/adapter/presenter", "CuiPresenter")
	webP := countImplTypes(t, "internal/adapter/presenter", "WebPresenter")

	data, err := os.ReadFile(filepath.Join(repoRoot, "docs/design/backend.md"))
	if err != nil {
		t.Fatalf("read backend.md: %v", err)
	}
	doc := string(data)

	cases := []struct {
		name string
		re   *regexp.Regexp
		want []int
	}{
		{
			"controller implementation types",
			regexp.MustCompile(`実装型は (\d+) 種類 \(CuiController (\d+) \+ WebController (\d+)\)`),
			[]int{cui + web, cui, web},
		},
		{
			"presenter implementation types",
			regexp.MustCompile(`実装型は (\d+) 種類 \(CuiPresenter (\d+) \+ WebPresenter (\d+)\)`),
			[]int{cuiP + webP, cuiP, webP},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			matches := c.re.FindAllStringSubmatch(doc, -1)
			if len(matches) == 0 {
				t.Fatalf("backend.md: pattern %q matched nothing -- the wording changed; update the regex here", c.re)
			}
			for _, m := range matches {
				for i, want := range c.want {
					if m[i+1] != fmt.Sprint(want) {
						t.Errorf("backend.md: %q says %s where the code has %d", m[0], m[i+1], want)
					}
				}
			}
		})
	}
}
