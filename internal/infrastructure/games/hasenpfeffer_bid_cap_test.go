package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestHasenpfefferBidMaxMatchesTheDomain guards the ceiling the Web quotes.
//
// The CUI names the cap from `domain.HasenpfefferMaxBid` while the page has its
// own `BID_MAX` literal, used both to build the bid buttons and (now) to explain
// why there are none (#5758). A drift would leave the page offering — or
// refusing — a bid the server disagrees about, with every existing test passing.
func TestHasenpfefferBidMaxMatchesTheDomain(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "pages", "HasenpfefferPage.tsx")
	src, err := os.ReadFile(path) //nolint:gosec // test-only, fixed path
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	m := regexp.MustCompile(`const BID_MAX = (\d+);`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("HasenpfefferPage.tsx no longer states BID_MAX as a literal")
	}
	got, err := strconv.Atoi(m[1])
	if err != nil {
		t.Fatalf("parse BID_MAX: %v", err)
	}
	if got != domain.HasenpfefferMaxBid {
		t.Errorf("the page caps bidding at %d, the domain at %d", got, domain.HasenpfefferMaxBid)
	}
}
