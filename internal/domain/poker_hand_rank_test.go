//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// #5595: 両画面が「ワンペア以上、または A と K を含む手で成立」と書くようになった。
// **その文が実際の判定と一致していること**を、総当たりで確かめる (受け入れ条件3)。
// 文言そのものは機械的に比べられないが、文が主張する規則は比べられる。
func TestDealerQualifies_MatchesTheRuleTheUiStates(t *testing.T) {
	suits := []int{CardDesignSpade, CardDesignHeart, CardDesignClover, CardDesignDiamond}
	values := []int{1, 2, 5, 9, 12, 13}

	stated := func(rank int, hand []*Card) bool {
		if rank >= PokerHandOnePair {
			return true
		}
		ace, king := false, false
		for _, c := range hand {
			if c.GetValue() == 1 {
				ace = true
			}
			if c.GetValue() == 13 {
				king = true
			}
		}
		return ace && king
	}

	checked := 0
	// 5 枚の手を総当たりに近い形で作る (スート 4 × ランク 6 から重複なしで選ぶ)。
	var hand []*Card
	var walk func(start int)
	walk = func(start int) {
		if len(hand) == 5 {
			checked++
			rank := evalFiveCardHand(hand)
			assert.Equal(t, stated(rank, hand), dealerQualifies(rank, hand),
				"hand %v (rank %d)", hand, rank)
			return
		}
		for i := start; i < len(suits)*len(values); i++ {
			c := NewCard(suits[i/len(values)], values[i%len(values)], true)
			hand = append(hand, c)
			walk(i + 1)
			hand = hand[:len(hand)-1]
		}
	}
	walk(0)

	// 0 件で成功と読まれないこと。
	assert.Greater(t, checked, 1000, "the sweep must actually build hands")
}
