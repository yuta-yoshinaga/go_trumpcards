//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// **四色牌は 4 色 × 7 種 × 4 枚 = 112 枚。** 52 枚デッキではない。
func TestTuSacDeck_Composition(t *testing.T) {
	deck := buildTuSacDeck()
	require.Len(t, deck, TuSacDeckSize)
	assert.Equal(t, 112, TuSacDeckSize, "四色牌は 112 枚")

	counts := map[[2]int]int{}
	for _, c := range deck {
		counts[[2]int{c.GetDesign(), c.GetValue()}]++
	}
	assert.Len(t, counts, TuSacColorCount*TuSacPieceCount, "色 × 種の組合せ数が合わない")
	for k, n := range counts {
		assert.Equal(t, TuSacCopies, n, "色 %d 駒 %d の枚数", k[0], k[1])
	}
}

// **車・馬・砲だけが色をまたげる。** 広げると別のゲームになる。
func TestTuSacIsChariotHorseCannon(t *testing.T) {
	for v := TuSacPieceMin; v <= TuSacPieceMax; v++ {
		want := v == TuSacPieceChariot || v == TuSacPieceHorse || v == TuSacPieceCannon
		assert.Equal(t, want, TuSacIsChariotHorseCannon(v), "駒 %d (%s)", v, TuSacPieceName(v))
	}
}

func TestTuSacNames(t *testing.T) {
	seen := map[string]bool{}
	for d := TuSacColorMin; d <= TuSacColorMax; d++ {
		n := TuSacColorName(d)
		assert.NotEqual(t, "unknown", n)
		assert.False(t, seen[n], "色の名前が重複: %s", n)
		seen[n] = true
	}
	seen = map[string]bool{}
	for v := TuSacPieceMin; v <= TuSacPieceMax; v++ {
		n := TuSacPieceName(v)
		assert.NotEqual(t, "unknown", n)
		assert.False(t, seen[n], "駒の名前が重複: %s", n)
		seen[n] = true
	}
	assert.Equal(t, "unknown", TuSacColorName(99))
	assert.Equal(t, "unknown", TuSacPieceName(99))
}
