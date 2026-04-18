package games

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/ui"
)

// expected category counts derived from the Phase 2 design: the three
// Cloudflare Workers split 55 games into casino (18) / classic (21) / solo
// (16). A mismatch here indicates that a game's Category is wrong (and would
// route to the wrong worker in production).
const (
	expectedCasino  = 18
	expectedClassic = 21
	expectedSolo    = 16
	expectedTotal   = expectedCasino + expectedClassic + expectedSolo
)

func TestAllReturnsExpectedTotal(t *testing.T) {
	if got := len(All()); got != expectedTotal {
		t.Fatalf("len(All()) = %d, want %d", got, expectedTotal)
	}
}

// TestAllReturnsIndependentCopy verifies that mutations to the returned
// slice — at both slice and field level — do not leak back into the global
// registry. Both dimensions matter: a shared backing array lets a caller
// overwrite neighbouring entries, and shared struct pointers let a caller
// corrupt a game's Name or factories.
func TestAllReturnsIndependentCopy(t *testing.T) {
	a := All()
	if len(a) == 0 {
		t.Fatal("All() returned empty slice")
	}
	originalName := a[0].Name
	a[0].Name = "mutated"

	b := All()
	if b[0].Name != originalName {
		t.Fatalf("mutation leaked: b[0].Name = %q, want %q", b[0].Name, originalName)
	}
}

func TestByCategoryCounts(t *testing.T) {
	cases := []struct {
		cat  Category
		want int
	}{
		{CategoryCasino, expectedCasino},
		{CategoryClassic, expectedClassic},
		{CategorySolo, expectedSolo},
	}
	for _, c := range cases {
		t.Run(c.cat.String(), func(t *testing.T) {
			if got := len(ByCategory(c.cat)); got != c.want {
				t.Fatalf("ByCategory(%s) = %d, want %d", c.cat, got, c.want)
			}
		})
	}
}

func TestCategoryString(t *testing.T) {
	cases := map[Category]string{
		CategoryCasino:  "casino",
		CategoryClassic: "classic",
		CategorySolo:    "solo",
		Category(99):    "Category(99)",
	}
	for cat, want := range cases {
		if got := cat.String(); got != want {
			t.Errorf("Category(%d).String() = %q, want %q", int(cat), got, want)
		}
	}
}

func TestAllEntriesAreValid(t *testing.T) {
	seen := make(map[string]bool, expectedTotal)
	for i, g := range All() {
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
		case CategoryCasino, CategoryClassic, CategorySolo:
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
	all := All()
	if len(cliNames) != len(all) {
		t.Fatalf("count mismatch: CLI=%d, registry=%d", len(cliNames), len(all))
	}
	for i, name := range cliNames {
		if all[i].Name != name {
			t.Errorf("order mismatch at index %d: CLI=%q, registry=%q", i, name, all[i].Name)
		}
	}
}
