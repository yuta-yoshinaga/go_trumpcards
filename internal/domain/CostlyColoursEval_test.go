//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ccCard(design, value int) *Card { return NewCard(design, value, true) }

func TestCostlyCardValue(t *testing.T) {
	assert.Equal(t, 1, CostlyCardValue(ccCard(CardDesignSpade, 1)), "A は 1")
	assert.Equal(t, 7, CostlyCardValue(ccCard(CardDesignHeart, 7)))
	for _, v := range []int{10, 11, 12, 13} {
		assert.Equal(t, 10, CostlyCardValue(ccCard(CardDesignClover, v)), "%d は 10", v)
	}
	assert.Equal(t, 0, CostlyCardValue(nil))
}

func TestCostlyIsJackOrDeuce(t *testing.T) {
	assert.True(t, CostlyIsJackOrDeuce(ccCard(CardDesignSpade, 11)))
	assert.True(t, CostlyIsJackOrDeuce(ccCard(CardDesignHeart, 2)))
	// **特別なのは J と 2 だけ。** Q や K は普通の札。
	for _, v := range []int{1, 3, 10, 12, 13} {
		assert.False(t, CostlyIsJackOrDeuce(ccCard(CardDesignClover, v)), "%d が特別扱いされている", v)
	}
}

// **25 が Cribbage との分かれ目。** 15 と 31 だけにすると、この派生の
// 特徴がひとつ消える。
func TestCostlyPlayScore_FifteenTwentyFiveThirtyOne(t *testing.T) {
	for _, tc := range []struct {
		name   string
		total  int
		pile   []*Card
		want   int
		reason string
	}{
		{"fifteen", 15, []*Card{ccCard(CardDesignSpade, 7), ccCard(CardDesignHeart, 8)}, 2, "fifteen"},
		{"twenty-five", 25, []*Card{
			ccCard(CardDesignSpade, 10), ccCard(CardDesignHeart, 10), ccCard(CardDesignClover, 5),
		}, 3, "twentyFive"},
		{"thirty-one", 31, []*Card{
			ccCard(CardDesignSpade, 10), ccCard(CardDesignHeart, 10), ccCard(CardDesignClover, 8),
			ccCard(CardDesignDiamond, 3),
		}, 4, "thirtyOne"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pts, reasons := CostlyPlayScore(tc.pile, tc.total)
			assert.Equal(t, tc.want, pts, "作った枚数ぶん点になっていない")
			assert.Contains(t, reasons, tc.reason)
		})
	}

	// 節目でなければ 0。
	pts, reasons := CostlyPlayScore([]*Card{ccCard(CardDesignSpade, 7)}, 7)
	assert.Equal(t, 0, pts)
	assert.Empty(t, reasons)
}

// **プライアルは 9、ダブルプライアルは 18。** Cribbage の 6 / 12 ではない。
func TestCostlyPlayScore_PairsAndPrials(t *testing.T) {
	pair := []*Card{ccCard(CardDesignSpade, 7), ccCard(CardDesignHeart, 7)}
	pts, reasons := CostlyPlayScore(pair, 14)
	assert.Equal(t, 2, pts)
	assert.Contains(t, reasons, "pair")

	prial := append(pair, ccCard(CardDesignClover, 7))
	pts, reasons = CostlyPlayScore(prial, 21)
	assert.Equal(t, 9, pts, "プライアルが 9 点になっていない")
	assert.Contains(t, reasons, "prial")

	double := append(prial, ccCard(CardDesignDiamond, 7))
	pts, reasons = CostlyPlayScore(double, 28)
	assert.Equal(t, 18, pts, "ダブルプライアルが 18 点になっていない")
	assert.Contains(t, reasons, "doublePrial")

	// リテラルで固定する (定数どうしの比較は実験にならない)。
	assert.Equal(t, 9, CostlyPrialPoints)
	assert.Equal(t, 18, CostlyDoublePrialPoints)
}

// **階段は並び順を問わない。** 5-7-6 と出ても 3 枚として数える。
func TestCostlyRunLength(t *testing.T) {
	assert.Equal(t, 3, CostlyRunLength([]*Card{
		ccCard(CardDesignSpade, 5), ccCard(CardDesignHeart, 7), ccCard(CardDesignClover, 6),
	}), "順不同の階段が数えられていない")
	assert.Equal(t, 4, CostlyRunLength([]*Card{
		ccCard(CardDesignSpade, 5), ccCard(CardDesignHeart, 7), ccCard(CardDesignClover, 6),
		ccCard(CardDesignDiamond, 8),
	}))
	// 2 枚は階段ではない。
	assert.Equal(t, 0, CostlyRunLength([]*Card{
		ccCard(CardDesignSpade, 5), ccCard(CardDesignHeart, 6),
	}))
	// 同位が混ざれば階段ではない。
	assert.Equal(t, 0, CostlyRunLength([]*Card{
		ccCard(CardDesignSpade, 5), ccCard(CardDesignHeart, 5), ccCard(CardDesignClover, 6),
	}))
	// 飛んでいれば階段ではない。
	assert.Equal(t, 0, CostlyRunLength([]*Card{
		ccCard(CardDesignSpade, 5), ccCard(CardDesignHeart, 7), ccCard(CardDesignClover, 9),
	}))
}

// **色とスートの梯子。** 一番上の「4 枚同スート」がゲーム名の由来。
func TestCostlyColourCombo(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cards []*Card
		combo string
		pts   int
	}{
		{"four in suit is Costly Colours", []*Card{
			ccCard(CardDesignSpade, 3), ccCard(CardDesignSpade, 5),
			ccCard(CardDesignSpade, 9), ccCard(CardDesignSpade, 12),
		}, CostlyComboCostlyColours, 6},
		{"four in colour, three in suit", []*Card{
			ccCard(CardDesignHeart, 3), ccCard(CardDesignHeart, 5),
			ccCard(CardDesignHeart, 9), ccCard(CardDesignDiamond, 12),
		}, CostlyComboFourInColourThree, 5},
		{"four in colour, two in suit", []*Card{
			ccCard(CardDesignHeart, 3), ccCard(CardDesignHeart, 5),
			ccCard(CardDesignDiamond, 9), ccCard(CardDesignDiamond, 12),
		}, CostlyComboFourInColourTwo, 4},
		{"three in suit", []*Card{
			ccCard(CardDesignSpade, 3), ccCard(CardDesignSpade, 5),
			ccCard(CardDesignSpade, 9), ccCard(CardDesignHeart, 12),
		}, CostlyComboThreeInSuit, 3},
		{"three in colour", []*Card{
			ccCard(CardDesignSpade, 3), ccCard(CardDesignSpade, 5),
			ccCard(CardDesignClover, 9), ccCard(CardDesignHeart, 12),
		}, CostlyComboThreeInColour, 2},
		{"nothing", []*Card{
			ccCard(CardDesignSpade, 3), ccCard(CardDesignSpade, 5),
			ccCard(CardDesignHeart, 9), ccCard(CardDesignDiamond, 12),
		}, CostlyComboNone, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			combo, pts := CostlyColourCombo(tc.cards)
			assert.Equal(t, tc.combo, combo)
			assert.Equal(t, tc.pts, pts)
		})
	}

	// 4 枚でなければ役にならない (ショーは手札 3 + 表の 1)。
	_, pts := CostlyColourCombo([]*Card{ccCard(CardDesignSpade, 3), ccCard(CardDesignSpade, 5)})
	assert.Equal(t, 0, pts)

	// 梯子はリテラルで固定する。
	assert.Equal(t, []int{2, 3, 4, 5, 6}, []int{
		CostlyThreeInColourPoints, CostlyThreeInSuitPoints,
		CostlyFourInColourTwoSuitPoints, CostlyFourInColourThreeSuitPoints, CostlyColoursPoints,
	})
}

// **J と 2 はトランプなら 4、それ以外は 2。**
func TestCostlyJackDeucePoints(t *testing.T) {
	hand := []*Card{
		ccCard(CardDesignSpade, 11), // トランプが ♠ なら 4
		ccCard(CardDesignHeart, 2),  // それ以外なので 2
		ccCard(CardDesignClover, 7), // 対象外
	}
	assert.Equal(t, 6, CostlyJackDeucePoints(hand, CardDesignSpade), "♠J 4 + ♥2 2 = 6")
	// 2 枚ともトランプ以外なら 4。
	assert.Equal(t, 4, CostlyJackDeucePoints(hand, CardDesignDiamond))
	// 対象が無ければ 0。
	assert.Equal(t, 0, CostlyJackDeucePoints([]*Card{ccCard(CardDesignSpade, 7)}, CardDesignSpade))
}

// **J と 2 は同位役では数えない。** 別枠で点になるので二重に数えない。
func TestCostlyRankCombo_SkipsJacksAndDeuces(t *testing.T) {
	n, pts := CostlyRankCombo([]*Card{
		ccCard(CardDesignSpade, 11), ccCard(CardDesignHeart, 11),
		ccCard(CardDesignClover, 7), ccCard(CardDesignDiamond, 3),
	})
	assert.Equal(t, 0, n, "J のペアが同位役として数えられている")
	assert.Equal(t, 0, pts)

	n, pts = CostlyRankCombo([]*Card{
		ccCard(CardDesignSpade, 7), ccCard(CardDesignHeart, 7),
		ccCard(CardDesignClover, 7), ccCard(CardDesignDiamond, 3),
	})
	assert.Equal(t, 3, n)
	assert.Equal(t, 9, pts, "プライアルが 9 点になっていない")
}

func TestCostlyIsRed(t *testing.T) {
	assert.True(t, CostlyIsRed(ccCard(CardDesignHeart, 5)))
	assert.True(t, CostlyIsRed(ccCard(CardDesignDiamond, 5)))
	assert.False(t, CostlyIsRed(ccCard(CardDesignSpade, 5)))
	assert.False(t, CostlyIsRed(ccCard(CardDesignClover, 5)))
	assert.False(t, CostlyIsRed(nil))
}

// **梯子の点は互いに違う。** 同じ点が並ぶと、どの役を取ったのか区別できない。
func TestCostlyColourLadderIsStrictlyIncreasing(t *testing.T) {
	ladder := []int{
		CostlyThreeInColourPoints, CostlyThreeInSuitPoints,
		CostlyFourInColourTwoSuitPoints, CostlyFourInColourThreeSuitPoints, CostlyColoursPoints,
	}
	for i := 1; i < len(ladder); i++ {
		require.Greater(t, ladder[i], ladder[i-1], "梯子の %d 段目が上がっていない", i)
	}
}
