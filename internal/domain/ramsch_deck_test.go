//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// **ラムシュの卓に配られるのはスカート牌 32 枚。** 得点計算も序列も
// 値 {A,7,8,9,10,J,Q,K} を前提に書かれているので、2〜6 が混ざると
// 「0 点で最弱の札」が静かに増え、120 点という総和も崩れる。
//
// スカートは**ダイヤ 0 枚**で出荷されたことがある (#5296)。枚数だけの assert は
// それを通すので、スートごとの枚数と値の集合まで見る。
func TestRamsch_DealsTheGerman32CardPack(t *testing.T) {
	g := domain.NewDefaultRamsch()
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
	// Reset は人間の手番まで CPU を進めるので、既に場に出ている札も数える。
	for _, tc := range g.GetCurrentTrick() {
		count(tc.Card)
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
		assert.Zero(t, values[v], "ラムシュに入らない値 %d が配られている", v)
	}
}
