//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func s27(design, value int) *Card { return NewCard(design, value, false) }

// **絵札は 0.5 点。** ここが Seven Twenty-Seven 固有の値で、×2 の整数表現の
// 存在理由そのもの。float64 に戻されたらここが最初に壊れる。
func TestSevenTwentySeven_CardValues(t *testing.T) {
	for _, tt := range []struct {
		name  string
		card  *Card
		want  int
		human string
	}{
		{"two", s27(CardDesignSpade, 2), 4, "2"},
		{"nine", s27(CardDesignHeart, 9), 18, "9"},
		{"ten", s27(CardDesignClover, 10), 20, "10"},
		{"jack is a half", s27(CardDesignSpade, 11), 1, "0.5"},
		{"queen is a half", s27(CardDesignHeart, 12), 1, "0.5"},
		{"king is a half", s27(CardDesignDiamond, 13), 1, "0.5"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sevenTwentySevenCardValue(tt.card))
			assert.Equal(t, tt.human, SevenTwentySevenFormat(tt.want))
		})
	}
	// エースは単独では決まらない。手全体で 1 か 11 を選ぶ。
	assert.Equal(t, 0, sevenTwentySevenCardValue(s27(CardDesignSpade, 1)))
}

// **エースは 1 枚ごとに 1 か 11。** どのエースを高くするかは合計に効かないので、
// 「11 として数える枚数」だけを 0..n で回せば全通り出る。
func TestSevenTwentySeven_AcesGiveEveryCombination(t *testing.T) {
	// A A 5 → 1+1+5=7(14), 1+11+5=17(34), 11+11+5=27(54)
	hand := []*Card{s27(CardDesignSpade, 1), s27(CardDesignHeart, 1), s27(CardDesignClover, 5)}
	assert.Equal(t, []int{14, 34, 54}, sevenTwentySevenTotals(hand))

	// 絵札 3 枚 = 1.5 点。エース無しなら選択肢は 1 つだけ。
	faces := []*Card{s27(CardDesignSpade, 11), s27(CardDesignHeart, 12), s27(CardDesignClover, 13)}
	assert.Equal(t, []int{3}, sevenTwentySevenTotals(faces))
	assert.Equal(t, "1.5", SevenTwentySevenFormat(3))
}

// **両方の目標を同時に狙える手がある。** A-A-5 は 7 ちょうどにも 27 ちょうどにも
// できる ── このゲームの華で、片側だけ見る実装では出せない。
func TestSevenTwentySeven_OneHandCanHitBothTargets(t *testing.T) {
	hand := []*Card{s27(CardDesignSpade, 1), s27(CardDesignHeart, 1), s27(CardDesignClover, 5)}

	low, okLow := SevenTwentySevenBestFor(hand, SevenTwentySevenLowTarget)
	require.True(t, okLow)
	assert.Equal(t, SevenTwentySevenLowTarget, low, "7 ちょうどにできる")

	high, okHigh := SevenTwentySevenBestFor(hand, SevenTwentySevenHighTarget)
	require.True(t, okHigh)
	assert.Equal(t, SevenTwentySevenHighTarget, high, "27 ちょうどにできる")
}

// **超えたらその側は失格。** 低い側だけ落ちて高い側は生きている、という状態が
// 普通に起きる。片方の判定を共有すると、この非対称が消える。
func TestSevenTwentySeven_BustingOneSideLeavesTheOtherAlive(t *testing.T) {
	// 10 + 9 = 19 点。7 は超えているが 27 は超えていない。
	hand := []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 9)}

	_, okLow := SevenTwentySevenBestFor(hand, SevenTwentySevenLowTarget)
	assert.False(t, okLow, "19 点で 7 側が生き残っている")

	high, okHigh := SevenTwentySevenBestFor(hand, SevenTwentySevenHighTarget)
	require.True(t, okHigh, "19 点で 27 側まで失格になっている")
	assert.Equal(t, 38, high)

	// 両方超える手。
	both := []*Card{s27(CardDesignSpade, 10), s27(CardDesignHeart, 10), s27(CardDesignClover, 10)}
	_, okLow2 := SevenTwentySevenBestFor(both, SevenTwentySevenLowTarget)
	_, okHigh2 := SevenTwentySevenBestFor(both, SevenTwentySevenHighTarget)
	assert.False(t, okLow2)
	assert.False(t, okHigh2, "30 点で 27 側が生き残っている")
}

// **目標を超えない範囲で最大を採る。** 「近い」は「小さいほうに近い」ではなく
// 「超えずに大きい」。A を低く数えて遠ざかるのは負け筋。
func TestSevenTwentySeven_PicksTheHighestTotalUnderTheTarget(t *testing.T) {
	// A + 5 → 6(12) か 16(32)。7 側は 6 を採る（16 は超過）。
	hand := []*Card{s27(CardDesignSpade, 1), s27(CardDesignHeart, 5)}
	low, ok := SevenTwentySevenBestFor(hand, SevenTwentySevenLowTarget)
	require.True(t, ok)
	assert.Equal(t, 12, low, "6 点になるはず")
	assert.Equal(t, "6", SevenTwentySevenFormat(low))

	// 27 側は 16 を採る（6 より 27 に近い）。
	high, okHigh := SevenTwentySevenBestFor(hand, SevenTwentySevenHighTarget)
	require.True(t, okHigh)
	assert.Equal(t, 32, high, "16 点になるはず")
}

// 全部の値を総当たりで突き合わせるオラクル。**倍率表現が実点と一致していること。**
func TestSevenTwentySeven_ScaledValuesMatchTheRealPoints(t *testing.T) {
	realPoints := map[int]float64{
		1: 0, // エースは別扱い
		2: 2, 3: 3, 4: 4, 5: 5, 6: 6, 7: 7, 8: 8, 9: 9, 10: 10,
		11: 0.5, 12: 0.5, 13: 0.5,
	}
	for v := 1; v <= 13; v++ {
		got := sevenTwentySevenCardValue(s27(CardDesignSpade, v))
		want := int(realPoints[v] * SevenTwentySevenScoreScale)
		assert.Equal(t, want, got, "value %d の内部点", v)
	}
	// 目標も実点と一致していること。
	assert.Equal(t, "7", SevenTwentySevenFormat(SevenTwentySevenLowTarget))
	assert.Equal(t, "27", SevenTwentySevenFormat(SevenTwentySevenHighTarget))
}
