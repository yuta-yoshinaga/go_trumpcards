//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDoubleExposureTestBlackJack(t *testing.T) *BlackJack {
	t.Helper()
	bj := NewDefaultBlackJack()
	cfg := bj.GetConfig()
	cfg.Variant = BJVariantDoubleExposure
	require.NoError(t, bj.SetConfig(cfg))
	return bj
}

func doubleExposureHand(cards ...*Card) *BlackJackHand {
	hand := NewBlackJackHand()
	for _, card := range cards {
		hand.AddCard(card)
	}
	hand.SetBet(100)
	return hand
}

// TestDoubleExposureVariantConfig verifies the variant knobs and JSON resolver.
func TestDoubleExposureVariantConfig(t *testing.T) {
	v := DoubleExposureVariant()
	require.NotNil(t, v)
	assert.Equal(t, BJVariantDoubleExposure, v.Name)
	assert.True(t, v.DealerCardsFaceUp)
	assert.True(t, v.DealerWinsTies)
	assert.True(t, v.BlackjackPaysEven)
	assert.True(t, v.PlayerBJBeatsDealerBJ)
	assert.Nil(t, v.DeckBuilder)

	resolved := ResolveBlackJackVariant(BJVariantDoubleExposure)
	require.NotNil(t, resolved)
	assert.Equal(t, v.Name, resolved.Name)
}

// TestDoubleExposureDealerWinsOrdinaryTies verifies that an equal non-BJ score loses.
func TestDoubleExposureDealerWinsOrdinaryTies(t *testing.T) {
	bj := newDoubleExposureTestBlackJack(t)
	hand := doubleExposureHand(NewCard(CardDesignSpade, 10, false), NewCard(CardDesignHeart, 8, false))
	bj.SetPlayerHands([]*BlackJackHand{hand})
	bj.GetDealer().AddCard(NewCard(CardDesignClover, 10, false))
	bj.GetDealer().AddCard(NewCard(CardDesignDiamond, 8, false))

	assert.Equal(t, GameResultLose, bj.GameJudgment())
}

// TestDoubleExposureNaturalBJBeatsDealerNaturalBJ verifies the tie exception.
func TestDoubleExposureNaturalBJBeatsDealerNaturalBJ(t *testing.T) {
	bj := newDoubleExposureTestBlackJack(t)
	hand := doubleExposureHand(NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 13, false))
	bj.SetPlayerHands([]*BlackJackHand{hand})
	bj.GetDealer().AddCard(NewCard(CardDesignClover, 1, false))
	bj.GetDealer().AddCard(NewCard(CardDesignDiamond, 13, false))

	assert.Equal(t, GameResultWin, bj.GameJudgment())
}

// TestDoubleExposureBlackjackPaysEven verifies the natural BJ 1:1 payout.
func TestDoubleExposureBlackjackPaysEven(t *testing.T) {
	bj := newDoubleExposureTestBlackJack(t)
	bj.GetPlayer().SetChips(900)
	hand := doubleExposureHand(NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 13, false))

	bj.payoutHandWithVariant(bj.GetPlayer(), hand, false, GameResultWin)

	assert.Equal(t, 1100, bj.GetPlayer().GetChips())
}

// TestDoubleExposureBustStillLoses verifies that busts remain immediate losses.
func TestDoubleExposureBustStillLoses(t *testing.T) {
	bj := newDoubleExposureTestBlackJack(t)
	hand := doubleExposureHand(NewCard(CardDesignSpade, 10, false), NewCard(CardDesignHeart, 8, false), NewCard(CardDesignClover, 5, false))
	bj.SetPlayerHands([]*BlackJackHand{hand})
	bj.GetDealer().AddCard(NewCard(CardDesignDiamond, 10, false))
	bj.GetDealer().AddCard(NewCard(CardDesignSpade, 7, false))

	assert.Equal(t, GameResultLose, bj.GameJudgment())
}

// TestDoubleExposureNilVariantRestoresStandardRules is the negative control.
func TestDoubleExposureNilVariantRestoresStandardRules(t *testing.T) {
	bj := NewDefaultBlackJack()
	hand := doubleExposureHand(NewCard(CardDesignSpade, 10, false), NewCard(CardDesignHeart, 8, false))
	bj.SetPlayerHands([]*BlackJackHand{hand})
	bj.GetDealer().AddCard(NewCard(CardDesignClover, 10, false))
	bj.GetDealer().AddCard(NewCard(CardDesignDiamond, 8, false))
	assert.Equal(t, GameResultDraw, bj.GameJudgment())

	bj.GetPlayer().SetChips(900)
	bjHand := doubleExposureHand(NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 13, false))
	bj.payoutHandWithVariant(bj.GetPlayer(), bjHand, false, GameResultWin)
	assert.Equal(t, 1150, bj.GetPlayer().GetChips())
}
