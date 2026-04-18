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

func TestAllReturnsFreshCopy(t *testing.T) {
	a := All()
	b := All()
	if len(a) == 0 || len(b) == 0 {
		t.Fatal("All() returned empty slice")
	}
	// Mutating one copy must not affect the other.
	a[0] = nil
	if b[0] == nil {
		t.Fatal("All() returned a shared backing array; mutations leak between callers")
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
		if g == nil {
			t.Fatalf("entry %d is nil", i)
		}
		if g.Name == "" {
			t.Errorf("entry %d has empty Name", i)
		}
		if seen[g.Name] {
			t.Errorf("duplicate Name %q", g.Name)
		}
		seen[g.Name] = true
		if g.NewWebController == nil {
			t.Errorf("game %q has nil NewWebController", g.Name)
		} else {
			if ctrl := g.NewWebController(); ctrl == nil {
				t.Errorf("game %q: NewWebController() returned nil", g.Name)
			}
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
// the set registered by the CLI's gameRegistry. The CLI is the canonical
// source of the game catalog; any drift would mean CLI and Web/Worker disagree.
func TestRegistryMatchesCLI(t *testing.T) {
	cliNames := make(map[string]bool, len(ui.GameNames()))
	for _, name := range ui.GameNames() {
		cliNames[name] = true
	}
	webNames := make(map[string]bool, expectedTotal)
	for _, g := range All() {
		webNames[g.Name] = true
	}
	for name := range cliNames {
		if !webNames[name] {
			t.Errorf("CLI registers %q but games.registry does not", name)
		}
	}
	for name := range webNames {
		if !cliNames[name] {
			t.Errorf("games.registry includes %q but CLI does not", name)
		}
	}
}
