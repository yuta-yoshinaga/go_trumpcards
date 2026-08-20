package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestBlackHoleWebTotalCardsMatchesTheDomain guards the web page's copy of the
// deck size it counts progress against.
//
// `BlackHolePage.tsx` renders "swallowed n/52" from its own constant while the
// CUI renders the same line from `domain.BlackHoleTotalCards` (#5681). If the
// deck ever changes, the page would keep counting toward the old total.
func TestBlackHoleWebTotalCardsMatchesTheDomain(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "pages", "BlackHolePage.tsx")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	m := regexp.MustCompile(`const BLACKHOLE_TOTAL_CARDS = (\d+);`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("BlackHolePage.tsx no longer declares BLACKHOLE_TOTAL_CARDS")
	}
	if m[1] != strconv.Itoa(domain.BlackHoleTotalCards) {
		t.Errorf("BLACKHOLE_TOTAL_CARDS = %s, want %d", m[1], domain.BlackHoleTotalCards)
	}
}
