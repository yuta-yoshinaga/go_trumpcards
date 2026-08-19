package games_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// TestSlobberhannesQueenMatchesTheDomain guards the two copies of "which card
// is the penalty queen".
//
// The page warns while the queen is on the table (#5745) from
// `frontend/src/utils/slobberhannesQueen.ts`; the domain scores the penalty
// from `SlobberhannesQueenSuit` / `SlobberhannesQueenValue`. Both sides are
// internally consistent, so pointing the warning at the wrong card would show a
// calm board while the point is being lost.
func TestSlobberhannesQueenMatchesTheDomain(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "utils", "slobberhannesQueen.ts")
	src, err := os.ReadFile(path) //nolint:gosec // test-only, fixed path
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	design := regexp.MustCompile(`SLOBBERHANNES_QUEEN_DESIGN: Card\['design'\] = '(\w+)'`).FindStringSubmatch(string(src))
	value := regexp.MustCompile(`SLOBBERHANNES_QUEEN_VALUE = (\d+)`).FindStringSubmatch(string(src))
	if design == nil || value == nil {
		t.Fatal("slobberhannesQueen.ts no longer states the queen as literals")
	}

	// ドメイン側はテスト対象そのものなので、判定関数で突き合わせる。
	wantValue, err := strconv.Atoi(value[1])
	if err != nil {
		t.Fatalf("parse the queen value: %v", err)
	}
	card := domain.NewCard(webDesignToGo(t, design[1]), wantValue, true)
	if !domain.SlobberhannesIsPenaltyQueen(card) {
		t.Errorf("the frontend calls %s %d the penalty queen, the domain does not", design[1], wantValue)
	}
	// 負のコントロール: 隣のランクとスートは罰点札ではない。
	if domain.SlobberhannesIsPenaltyQueen(domain.NewCard(webDesignToGo(t, design[1]), wantValue-1, true)) {
		t.Error("the rank below must not count as the penalty queen")
	}
	if domain.SlobberhannesIsPenaltyQueen(domain.NewCard(domain.CardDesignSpade, wantValue, true)) {
		t.Error("the same rank in another suit must not count as the penalty queen")
	}
}

// webDesignToGo maps the frontend's design name onto the domain constant.
func webDesignToGo(t *testing.T, design string) int {
	t.Helper()
	switch design {
	case "SPADE":
		return domain.CardDesignSpade
	case "CLOVER":
		return domain.CardDesignClover
	case "HEART":
		return domain.CardDesignHeart
	case "DIAMOND":
		return domain.CardDesignDiamond
	}
	t.Fatalf("unknown design %q", design)
	return 0
}
