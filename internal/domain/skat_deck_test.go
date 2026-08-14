//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// **枚数だけの assert はこの不具合を通す。** 32 枚は合っていながら、
// ♠13 + ♣13 + ♥6 + ♦0 という「ダイヤが 1 枚も無い」デッキが配られていた。
// スートごとの枚数と値の集合まで見る。
func TestNewTrumpCards32_IsTheGerman32CardPack(t *testing.T) {
	want := map[int]bool{1: true, 7: true, 8: true, 9: true, 10: true, 11: true, 12: true, 13: true}

	for _, tt := range []struct {
		name string
		deck *domain.TrumpCards
	}{
		{"32", domain.NewTrumpCards32()},
		{"belote", domain.NewTrumpCardsBelote()},
		{"prsi", domain.NewTrumpCardsPrsi()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, 32, tt.deck.GetTotalCount())

			suits := map[int]int{}
			values := map[int]int{}
			for i := range 32 {
				c := tt.deck.DrawCard()
				require.NotNil(t, c, "index %d", i)
				assert.True(t, want[c.GetValue()], "値 %d はこのデッキに入らない", c.GetValue())
				suits[c.GetDesign()]++
				values[c.GetValue()]++
			}
			// **4 スートがそれぞれ 8 枚。** 切り札スートを選ぶゲームでは、
			// 1 スートでも欠けると「選べるのに 1 枚も無い切り札」ができる。
			for s := domain.CardDesignSpade; s <= domain.CardDesignDiamond; s++ {
				assert.Equal(t, 8, suits[s], "スート %d の枚数", s)
			}
			for v := range want {
				assert.Equal(t, 4, values[v], "値 %d の枚数", v)
			}
			assert.Nil(t, tt.deck.DrawCard(), "32 枚を超えて引けている")
		})
	}
}

// **スカートの卓に配られる 32 枚も同じ内容。** 得点計算 (skatCardPoints) も
// 序列も値 {A,7,8,9,10,J,Q,K} を前提に書かれているので、2〜6 が混ざると
// 「0 点で最弱の札」が静かに増える。
func TestSkat_DealsTheGerman32CardPack(t *testing.T) {
	g := domain.NewDefaultSkat()
	g.Reset()

	suits := map[int]int{}
	values := map[int]int{}
	count := func(c *domain.Card) {
		suits[c.GetDesign()]++
		values[c.GetValue()]++
	}
	for i := range g.GetPlayerCnt() {
		p := g.GetPlayer(i)
		for j := range p.GetCardsSize() {
			count(p.GetCard(j))
		}
	}
	for _, c := range g.GetSkat() {
		count(c)
	}

	total := 0
	for _, n := range suits {
		total += n
	}
	require.Equal(t, 32, total, "配られた枚数")
	for s := domain.CardDesignSpade; s <= domain.CardDesignDiamond; s++ {
		assert.Equal(t, 8, suits[s], "スート %d が配られていない", s)
	}
	for _, v := range []int{1, 7, 8, 9, 10, 11, 12, 13} {
		assert.Equal(t, 4, values[v], "値 %d の枚数", v)
	}
	for _, v := range []int{2, 3, 4, 5, 6} {
		assert.Zero(t, values[v], "スカートに入らない値 %d が配られている", v)
	}
}
