//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// newSettlementTable returns a standard table whose dealer stands on 17 without a blackjack,
// with one player hand built from the given cards and a 100-chip bet already taken.
func newSettlementTable(t *testing.T, playerCards ...*Card) (*BlackJack, *BlackJackHand) {
	t.Helper()
	bj := NewDefaultBlackJack()
	bj.GetPlayer().SetChips(900)
	bj.dealer.AddCard(NewCard(CardDesignSpade, 10, false))
	bj.dealer.AddCard(NewCard(CardDesignHeart, 7, false))
	hand := NewBlackJackHand()
	for _, c := range playerCards {
		hand.AddCard(c)
	}
	hand.SetBet(100)
	bj.playerHands = []*BlackJackHand{hand}
	return bj, hand
}

// The per-hand net change is what the payout actually credited minus the stake, so a
// natural pays 3:2 (+150) rather than the even-money +100 a rule re-derivation might give.
func TestBlackJackHandNetChange_FollowsActualPayout(t *testing.T) {
	t.Run("natural blackjack pays 3:2", func(t *testing.T) {
		bj, hand := newSettlementTable(t, NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 13, false))
		bj.resolvePayouts()
		assert.Equal(t, GameResultWin, hand.GetResult())
		assert.Equal(t, 150, hand.GetNetChange())
		assert.Equal(t, 900+250, bj.GetPlayer().GetChips())
	})

	t.Run("plain win pays even money", func(t *testing.T) {
		bj, hand := newSettlementTable(t, NewCard(CardDesignSpade, 10, false), NewCard(CardDesignHeart, 9, false))
		bj.resolvePayouts()
		assert.Equal(t, GameResultWin, hand.GetResult())
		assert.Equal(t, 100, hand.GetNetChange())
	})

	t.Run("loss forfeits the stake", func(t *testing.T) {
		bj, hand := newSettlementTable(t, NewCard(CardDesignSpade, 10, false), NewCard(CardDesignHeart, 6, false))
		bj.resolvePayouts()
		assert.Equal(t, GameResultLose, hand.GetResult())
		assert.Equal(t, -100, hand.GetNetChange())
	})

	t.Run("push returns the stake", func(t *testing.T) {
		bj, hand := newSettlementTable(t, NewCard(CardDesignSpade, 10, false), NewCard(CardDesignHeart, 7, false))
		bj.resolvePayouts()
		assert.Equal(t, GameResultDraw, hand.GetResult())
		assert.Equal(t, 0, hand.GetNetChange())
	})
}
