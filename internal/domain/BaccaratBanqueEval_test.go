//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bbCard(design, value int) *Card { return NewCard(design, value, true) }

// **10 と絵札は 0、A は 1。**
func TestBaccaratBanqueCardValue(t *testing.T) {
	assert.Equal(t, 1, BaccaratBanqueCardValue(bbCard(CardDesignSpade, 1)))
	assert.Equal(t, 9, BaccaratBanqueCardValue(bbCard(CardDesignHeart, 9)))
	for _, v := range []int{10, 11, 12, 13} {
		assert.Equal(t, 0, BaccaratBanqueCardValue(bbCard(CardDesignClover, v)), "%d が 0 でない", v)
	}
	assert.Equal(t, 0, BaccaratBanqueCardValue(nil))
}

// **合計は 10 で割った余り。**
func TestBaccaratBanqueTotal(t *testing.T) {
	assert.Equal(t, 5, BaccaratBanqueTotal([]*Card{
		bbCard(CardDesignSpade, 7), bbCard(CardDesignHeart, 8),
	}), "7+8=15 → 5")
	assert.Equal(t, 0, BaccaratBanqueTotal([]*Card{
		bbCard(CardDesignSpade, 10), bbCard(CardDesignHeart, 13),
	}), "絵札だけなら 0")
	assert.Equal(t, 9, BaccaratBanqueTotal([]*Card{
		bbCard(CardDesignSpade, 4), bbCard(CardDesignHeart, 5),
	}))
}

func TestBaccaratBanqueIsNatural(t *testing.T) {
	assert.True(t, BaccaratBanqueIsNatural([]*Card{
		bbCard(CardDesignSpade, 3), bbCard(CardDesignHeart, 5),
	}), "8 はナチュラル")
	assert.True(t, BaccaratBanqueIsNatural([]*Card{
		bbCard(CardDesignSpade, 4), bbCard(CardDesignHeart, 5),
	}), "9 はナチュラル")
	assert.False(t, BaccaratBanqueIsNatural([]*Card{
		bbCard(CardDesignSpade, 3), bbCard(CardDesignHeart, 4),
	}), "7 はナチュラルでない")
	// **3 枚ではナチュラルにならない。** 最初の 2 枚だけの話。
	assert.False(t, BaccaratBanqueIsNatural([]*Card{
		bbCard(CardDesignSpade, 3), bbCard(CardDesignHeart, 5), bbCard(CardDesignClover, 1),
	}))
}

// **裁量があるのは合計 5 のときだけ。** #5462 の冒頭は「合計 5 の子を除き
// 裁量」と裏返しに書いている ── 0-4 は必ず引き、6-7 は必ず止まる。
func TestBaccaratBanquePunterRule(t *testing.T) {
	// 合計 0〜4 は必ず引く。
	for total := 0; total <= 4; total++ {
		cards := bbHandTotalling(total)
		assert.Equal(t, BaccaratBanqueDrawMust, BaccaratBanquePunterRule(cards),
			"合計 %d で必ず引くことになっていない", total)
	}
	// 合計 5 だけが裁量。
	assert.Equal(t, BaccaratBanqueDrawFree, BaccaratBanquePunterRule(bbHandTotalling(5)),
		"合計 5 が裁量になっていない")
	// 合計 6〜7 は必ず止まる。
	for total := 6; total <= 7; total++ {
		assert.Equal(t, BaccaratBanqueDrawStand, BaccaratBanquePunterRule(bbHandTotalling(total)),
			"合計 %d で必ず止まることになっていない", total)
	}
	// 8〜9 はナチュラルなので 3 枚目が無い。
	for total := 8; total <= 9; total++ {
		assert.Equal(t, BaccaratBanqueDrawNatural, BaccaratBanquePunterRule(bbHandTotalling(total)),
			"合計 %d がナチュラルになっていない", total)
	}
}

// **バンカーはどの合計でも自由。** プント・バンコのような固定表は無い。
func TestBaccaratBanqueBankerRule(t *testing.T) {
	for total := 0; total <= 7; total++ {
		assert.Equal(t, BaccaratBanqueDrawFree, BaccaratBanqueBankerRule(bbHandTotalling(total)),
			"合計 %d でバンカーの裁量が失われている", total)
	}
	for total := 8; total <= 9; total++ {
		assert.Equal(t, BaccaratBanqueDrawNatural, BaccaratBanqueBankerRule(bbHandTotalling(total)))
	}
}

// **左右は別勘定。** 片方が勝ってもう片方が負けることがある。
func TestBaccaratBanqueCompare(t *testing.T) {
	assert.Equal(t, BaccaratBanqueOutcomeBankerWin, BaccaratBanqueCompare(7, 5))
	assert.Equal(t, BaccaratBanqueOutcomePunterWin, BaccaratBanqueCompare(4, 9))
	assert.Equal(t, BaccaratBanqueOutcomeTie, BaccaratBanqueCompare(6, 6))
}

// **シューは 3 組 156 枚。**
func TestNewBaccaratBanqueShoe(t *testing.T) {
	shoe := NewBaccaratBanqueShoe()
	require.Len(t, shoe, BaccaratBanqueDeckSize)
	assert.Equal(t, 156, BaccaratBanqueDeckSize, "3 組 156 枚でない")

	counts := map[[2]int]int{}
	for _, c := range shoe {
		counts[[2]int{c.GetDesign(), c.GetValue()}]++
	}
	require.Len(t, counts, 52, "52 種類でない")
	for key, n := range counts {
		assert.Equal(t, BaccaratBanqueDeckCount, n, "%v が %d 枚ある", key, n)
	}

	// **束ねたあと混ぜる。** 組ごとに並んだままだと、同じ札が決まった間隔で
	// 出てくる ── 先頭 52 枚に 52 種類が揃っていたら混ざっていない。
	interleaved := false
	first := map[[2]int]bool{}
	for _, c := range shoe[:52] {
		key := [2]int{c.GetDesign(), c.GetValue()}
		if first[key] {
			interleaved = true
			break
		}
		first[key] = true
	}
	assert.True(t, interleaved, "3 組が混ざっていない (先頭 52 枚が 1 組そのまま)")
}

// bbHandTotalling は合計が total になる 2 枚を返す。
func bbHandTotalling(total int) []*Card {
	// 10 (=0) と total で合計 total になる。
	return []*Card{bbCard(CardDesignSpade, 10), bbCard(CardDesignHeart, bbRankFor(total))}
}

// bbRankFor は点 v を持つランクを返す (0 は 10 で表す)。
func bbRankFor(v int) int {
	if v == 0 {
		return 10
	}
	return v
}
