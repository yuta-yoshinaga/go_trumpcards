//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cmtCard(design, value int) *Card { return NewCard(design, value, true) }

// **8♦ を抜いた 51 枚。** 抜いた札の位置で連なりが必ず止まる ── これが
// stops 系の名前の由来なので、52 枚のまま配ると止まりどころが消える。
func TestCometDeck(t *testing.T) {
	deck := NewCometDeck()
	require.Len(t, deck, 51)
	for _, c := range deck {
		assert.False(t, IsCometRemoved(c), "8♦ が残っている")
	}
	seen := map[[2]int]bool{}
	wild := 0
	for _, c := range deck {
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "同じ札が 2 枚ある: %v", key)
		seen[key] = true
		if IsCometWild(c) {
			wild++
		}
	}
	assert.Equal(t, 1, wild, "コメット (9♦) がちょうど 1 枚でない")
}

func TestIsCometWild(t *testing.T) {
	assert.True(t, IsCometWild(cmtCard(CardDesignDiamond, 9)))
	// **9 なら何でもワイルド、ではない。** ♦ の 9 だけ。
	assert.False(t, IsCometWild(cmtCard(CardDesignClover, 9)))
	assert.False(t, IsCometWild(cmtCard(CardDesignDiamond, 8)))
	assert.False(t, IsCometWild(nil))
}

// **#5459 の「同スート」は誤り。** コメットはスートを無視してランクだけで
// 繋ぐ ── そこが Michigan (Newmarket 系) と分かれる唯一の点で、同スートに
// すると Michigan の作り直しになる。
func TestCometSequenceIgnoresSuit(t *testing.T) {
	// 5 の次に要るのは 6。スートは問わない。
	for _, d := range []int{CardDesignSpade, CardDesignHeart, CardDesignClover, CardDesignDiamond} {
		assert.True(t, CanPlayComet(cmtCard(d, 6), 6), "スート %d の 6 が出せない", d)
	}
	assert.False(t, CanPlayComet(cmtCard(CardDesignSpade, 7), 6), "ランク違いが通っている")
	assert.False(t, CanPlayComet(cmtCard(CardDesignSpade, 5), 6))
}

// **連なりの先頭は何でもよい。** A 固定ではない。
func TestCometLeadIsAnyCard(t *testing.T) {
	for _, v := range []int{1, 7, 12, 13} {
		assert.True(t, CanPlayComet(cmtCard(CardDesignHeart, v), 0),
			"先頭に %d が出せない", v)
	}
}

// **コメットはどのランクの代わりにもなる。**
func TestCometWildSubstitutesForAnyRank(t *testing.T) {
	wild := cmtCard(CardDesignDiamond, 9)
	for need := CometMinRank; need <= CometMaxRank; need++ {
		assert.True(t, CanPlayComet(wild, need), "コメットが %d の代わりになれない", need)
	}
	assert.False(t, CanPlayComet(nil, 5))
}

// **K とコメットで連なりが切れる。** K は上限で次が無く、コメットは代役なので
// 次のランクが決まらない。
func TestCometStopsSequence(t *testing.T) {
	assert.True(t, CometStopsSequence(cmtCard(CardDesignSpade, 13)), "K で切れない")
	assert.True(t, CometStopsSequence(cmtCard(CardDesignDiamond, 9)), "コメットで切れない")
	assert.True(t, CometStopsSequence(nil))
	for _, v := range []int{1, 8, 12} {
		assert.False(t, CometStopsSequence(cmtCard(CardDesignHeart, v)),
			"%d で切れてしまっている", v)
	}
	// ♦ でない 9 はただの 9 なので切れない。
	assert.False(t, CometStopsSequence(cmtCard(CardDesignClover, 9)))
}

func TestCometPlayableIdxs(t *testing.T) {
	hand := []*Card{
		cmtCard(CardDesignSpade, 4),
		cmtCard(CardDesignHeart, 6),
		cmtCard(CardDesignDiamond, 9), // コメット
		cmtCard(CardDesignClover, 6),
	}
	// 6 が要るとき: ♥6・♣6・コメット。
	assert.Equal(t, []int{1, 2, 3}, CometPlayableIdxs(hand, 6))
	// 誰も持っていないランクでもコメットは出せる。
	assert.Equal(t, []int{2}, CometPlayableIdxs(hand, 11))
	// 先頭なら全部。
	assert.Len(t, CometPlayableIdxs(hand, 0), 4)
}

// **残した札は 1 枚 1 点。** 額面や絵札 10 点の方式は採らない。
func TestCometCardPoints(t *testing.T) {
	assert.Equal(t, 1, CometCardPoints(cmtCard(CardDesignSpade, 13)))
	assert.Equal(t, 1, CometCardPoints(cmtCard(CardDesignHeart, 2)))
	assert.Equal(t, 1, CometCardPoints(cmtCard(CardDesignDiamond, 9)))
}
