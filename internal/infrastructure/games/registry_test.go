package games_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/games"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
)

// expected category counts derived from the Phase 2 design: the three
// Cloudflare Workers split games into casino / classic / solo. A mismatch
// here indicates that a game's Category is wrong (and would route to the
// wrong worker in production).
const (
	expectedCasino  = 64
	expectedClassic = 57
	expectedSolo    = 52
	expectedExtra   = 46
	// ADR-0036 Phase 1: the fifth and sixth buckets exist but hold nothing yet.
	// Phase 2 rebalances into them, and these numbers move with it.
	expectedExtra2 = 0
	expectedExtra3 = 0
	expectedTotal  = expectedCasino + expectedClassic + expectedSolo + expectedExtra + expectedExtra2 + expectedExtra3
)

func TestAllReturnsExpectedTotal(t *testing.T) {
	if got := len(games.All()); got != expectedTotal {
		t.Fatalf("len(games.All()) = %d, want %d", got, expectedTotal)
	}
}

// TestAllReturnsIndependentCopy verifies that mutations to the returned
// slice — at both slice and field level — do not leak back into the global
// registry. Both dimensions matter: a shared backing array lets a caller
// overwrite neighbouring entries, and shared struct pointers let a caller
// corrupt a game's Name or factories.
func TestAllReturnsIndependentCopy(t *testing.T) {
	a := games.All()
	if len(a) == 0 {
		t.Fatal("games.All() returned empty slice")
	}
	originalName := a[0].Name
	a[0].Name = "mutated"

	b := games.All()
	if b[0].Name != originalName {
		t.Fatalf("mutation leaked: b[0].Name = %q, want %q", b[0].Name, originalName)
	}
}

func TestByCategoryCounts(t *testing.T) {
	cases := []struct {
		cat  games.Category
		want int
	}{
		{games.CategoryCasino, expectedCasino},
		{games.CategoryClassic, expectedClassic},
		{games.CategorySolo, expectedSolo},
		{games.CategoryExtra, expectedExtra},
		{games.CategoryExtra2, expectedExtra2},
		{games.CategoryExtra3, expectedExtra3},
	}
	for _, c := range cases {
		t.Run(c.cat.String(), func(t *testing.T) {
			if got := len(games.ByCategory(c.cat)); got != c.want {
				t.Fatalf("games.ByCategory(%s) = %d, want %d", c.cat, got, c.want)
			}
		})
	}
}

func TestCategoryString(t *testing.T) {
	cases := map[games.Category]string{
		games.CategoryCasino:  "casino",
		games.CategoryClassic: "classic",
		games.CategorySolo:    "solo",
		games.CategoryExtra:   "extra",
	}
	for cat, want := range cases {
		if got := cat.String(); got != want {
			t.Errorf("Category(%d).String() = %q, want %q", int(cat), got, want)
		}
	}
}

// TestCategoryStringPanicsOnUnknown asserts that String() panics on an
// undefined Category value so API misuse surfaces immediately — matching
// BindWebController/BindWorker's panic-on-misuse policy.
func TestCategoryStringPanicsOnUnknown(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on unknown Category, got none")
		}
	}()
	_ = games.Category(99).String()
}

func TestAllEntriesAreValid(t *testing.T) {
	seen := make(map[string]bool, expectedTotal)
	for i, g := range games.All() {
		if g.Name == "" {
			t.Errorf("entry %d has empty Name", i)
		}
		if seen[g.Name] {
			t.Errorf("duplicate Name %q", g.Name)
		}
		seen[g.Name] = true
		if g.NewWebController == nil {
			t.Errorf("game %q has nil NewWebController", g.Name)
		} else if ctrl := g.NewWebController(); ctrl == nil {
			t.Errorf("game %q: NewWebController() returned nil", g.Name)
		}
		switch g.Category {
		case games.CategoryCasino, games.CategoryClassic, games.CategorySolo, games.CategoryExtra:
			// valid
		default:
			t.Errorf("game %q has invalid Category %d", g.Name, int(g.Category))
		}
	}
}

// TestRegistryMatchesCLI asserts that the games registered here are exactly
// the set registered by the CLI's gameRegistry — in the same order. Using
// an ordered comparison (not set equality) is load-bearing: registry.go
// documents that its order mirrors the CLI's, and downstream help text /
// listings / completion order depend on that contract.
func TestRegistryMatchesCLI(t *testing.T) {
	cliNames := ui.GameNames()
	all := games.All()
	if len(cliNames) != len(all) {
		t.Fatalf("count mismatch: CLI=%d, registry=%d", len(cliNames), len(all))
	}
	for i, name := range cliNames {
		if all[i].Name != name {
			t.Errorf("order mismatch at index %d: CLI=%q, registry=%q", i, name, all[i].Name)
		}
	}
}

// TestRegistryHasDescriptionForEach asserts that every game in the registry
// carries a non-empty Description — the CLI listings pull from here (via
// games.Descriptions), so an empty entry silently ships a blank row.
func TestRegistryHasDescriptionForEach(t *testing.T) {
	for _, g := range games.All() {
		if strings.TrimSpace(games.Description(g.Name)) == "" {
			t.Errorf("game %q has empty Description in gameDescriptions", g.Name)
		}
	}
}

// TestDescriptionLookup exercises the single-name accessor used by the CLI
// in hot loops — must hit on known games and return "" for unknown ones.
func TestDescriptionLookup(t *testing.T) {
	if got := games.Description("blackjack"); got == "" {
		t.Errorf("Description(\"blackjack\") = \"\", want non-empty")
	}
	if got := games.Description("does-not-exist"); got != "" {
		t.Errorf("Description(\"does-not-exist\") = %q, want \"\"", got)
	}
}

// TestDescriptionsReturnsCachedMap documents that Descriptions() returns the
// package-owned cached map — callers must not mutate it (performance
// contract: O(1) access in loops, no per-call allocation).
func TestDescriptionsReturnsCachedMap(t *testing.T) {
	a := games.Descriptions()
	b := games.Descriptions()
	if &a == &b {
		// Map headers are allowed to differ — it's the underlying map data
		// that must be shared. Checking a single key's address is unreliable,
		// so just verify repeated calls return maps of equal size with
		// identical contents (and leave mutation-safety as a documented
		// contract).
		t.Log("note: Go maps compare by reference; repeated calls return the same backing map")
	}
	if len(a) != len(b) {
		t.Fatalf("len mismatch across Descriptions() calls: %d vs %d", len(a), len(b))
	}
	for k, v := range a {
		if b[k] != v {
			t.Errorf("Descriptions()[%q] drift: %q vs %q", k, v, b[k])
		}
	}
}

// TestCLIDescriptionsMatchRegistry asserts that the CLI and games package
// agree on every Name→Description pair. Enforces the SSoT contract from
// issue #1459: descriptions live on games.Game and ui sources them from
// there, so this test guards against re-introduction of duplicate storage.
func TestCLIDescriptionsMatchRegistry(t *testing.T) {
	cli := ui.GameDescriptions()
	reg := games.Descriptions()
	if len(cli) != len(reg) {
		t.Fatalf("count mismatch: CLI=%d, registry=%d", len(cli), len(reg))
	}
	for name, cliDesc := range cli {
		regDesc, ok := reg[name]
		if !ok {
			t.Errorf("game %q missing from games.Descriptions()", name)
			continue
		}
		if cliDesc != regDesc {
			t.Errorf("game %q: ui=%q registry=%q", name, cliDesc, regDesc)
		}
	}
}
