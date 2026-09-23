package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestDoubleKlondikeWebTotalCardsMatchesTheDomain guards the web page's copy of the
// deck size it counts progress against.
//
// `DoubleKlondikePage.tsx` renders "foundations n/104" from its own constant while the
// CUI renders the same line from `domain.DoubleKlondikeTotalCards` (#7339). If the
// deck ever changes, the page would keep counting toward the old total.
func TestDoubleKlondikeWebTotalCardsMatchesTheDomain(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "pages", "DoubleKlondikePage.tsx")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	m := regexp.MustCompile(`const DOUBLE_KLONDIKE_TOTAL_CARDS = (\d+);`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("DoubleKlondikePage.tsx no longer declares DOUBLE_KLONDIKE_TOTAL_CARDS")
	}
	if m[1] != strconv.Itoa(domain.DoubleKlondikeTotalCards) {
		t.Errorf("DOUBLE_KLONDIKE_TOTAL_CARDS = %s, want %d", m[1], domain.DoubleKlondikeTotalCards)
	}
}
