package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestConquianWebMeldTargetMatchesTheDomain guards the web page's copy of the
// number of melded cards that wins a round.
//
// `ConquianPage.tsx` draws its progress bar against `CONQUIAN_MELD_TARGET` and
// the CUI now prints the same progress from `domain.ConquianMeldTarget`. The
// figure is not arbitrary -- it is the deal size plus the card drawn to go out
// -- so a change to `ConquianHandSize` moves it, and a hardcoded 11 on the web
// would quietly keep counting to the old target (#5664).
func TestConquianWebMeldTargetMatchesTheDomain(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "pages", "ConquianPage.tsx")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	m := regexp.MustCompile(`const CONQUIAN_MELD_TARGET = (\d+);`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("ConquianPage.tsx no longer declares CONQUIAN_MELD_TARGET")
	}
	if m[1] != strconv.Itoa(domain.ConquianMeldTarget) {
		t.Errorf("CONQUIAN_MELD_TARGET = %s, want %d", m[1], domain.ConquianMeldTarget)
	}
	// 目標は配布枚数から導かれる。片方だけ動かすと進捗が意味を失う。
	if domain.ConquianMeldTarget != domain.ConquianHandSize+1 {
		t.Errorf("ConquianMeldTarget = %d, want ConquianHandSize+1 (%d)",
			domain.ConquianMeldTarget, domain.ConquianHandSize+1)
	}
}
