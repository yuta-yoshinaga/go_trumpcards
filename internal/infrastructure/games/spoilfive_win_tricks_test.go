package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestSpoilFiveWebWinTricksMatchesTheDomain guards the web page's copy of the
// "three tricks takes the round" threshold.
//
// `SpoilFivePage.tsx` renders every player's progress as `n/3` and flags the
// player one trick short, so it needs the number that `ResolveTrick` actually
// ends the round on (`SpoilFiveWinTricks`). Change the rule and the panel would
// keep counting toward the old target with nothing to catch it (#5655).
func TestSpoilFiveWebWinTricksMatchesTheDomain(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "pages", "SpoilFivePage.tsx")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	m := regexp.MustCompile(`const SPOILFIVE_WIN_TRICKS = (\d+);`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("SpoilFivePage.tsx no longer declares SPOILFIVE_WIN_TRICKS")
	}
	if m[1] != strconv.Itoa(domain.SpoilFiveWinTricks) {
		t.Errorf("SPOILFIVE_WIN_TRICKS = %s, want %d", m[1], domain.SpoilFiveWinTricks)
	}
}
