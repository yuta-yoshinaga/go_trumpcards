//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func deHand(values ...int) *BlackJackHand {
	h := NewBlackJackHand()
	for i, value := range values {
		design := CardDesignSpade
		if i%2 == 1 {
			design = CardDesignHeart
		}
		h.AddCard(NewCard(design, value, false))
	}
	return h
}

func deCard(value int) *Card {
	return NewCard(CardDesignClover, value, false)
}

func TestDoubleExposureStrategy_NegativeControls(t *testing.T) {
	hand := deHand(9, 3)
	up := deCard(4)
	hole := deCard(5)

	assert.Equal(t, BJSuggestStand, GetVariantStrategyAction(hand, up, hole, false, BJVariantStandard))
	assert.Equal(t, BJSuggestHit, GetVariantStrategyAction(hand, up, hole, false, BJVariantSpanish21))
	assert.Equal(t, BJSuggestStand, GetVariantStrategyAction(hand, up, nil, false, BJVariantDoubleExposure))
}

func TestDoubleExposureStrategy_DealerWinsNonNaturalTies(t *testing.T) {
	assert.NotEqual(t, BJSuggestStand, GetDoubleExposureStrategyAction(deHand(10, 10), 20, false))
	assert.NotEqual(t, BJSuggestStand, GetDoubleExposureStrategyAction(deHand(9, 8), 17, false))
}

func TestDoubleExposureStrategy_Hard20AgainstHard17Stands(t *testing.T) {
	// The generator must represent hard 20 as 10+10; the solver receives isPair=false.
	assert.Equal(t, newHand(10, 10), hardHandOfTotal(20))
	// Use three cards so the production lookup is deliberately a non-pair hard hand.
	hard20 := GetDoubleExposureStrategyAction(deHand(7, 7, 6), 17, false)
	hard12 := GetDoubleExposureStrategyAction(deHand(9, 3), 17, false)

	assert.Equal(t, BJSuggestStand, hard20)
	assert.NotEqual(t, hard12, hard20)
}

func TestDoubleExposureStrategy_NaturalAndSoftRules(t *testing.T) {
	assert.Equal(t, BJSuggestStand, GetDoubleExposureStrategyAction(deHand(1, 10), 20, false))
	assert.NotEqual(t,
		GetDoubleExposureStrategyAction(deHand(9, 3), 13, false),
		GetDoubleExposureStrategyAction(deHand(9, 3), 13, true),
	)
}

func TestDoubleExposureGame_SuggestionUsesVisibleDealerTotal(t *testing.T) {
	bj := NewDoubleExposureBlackJack()
	bj.ToggleHint()
	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	hand.AddCard(deCard(9))
	hand.AddCard(deCard(3))
	bj.GetDealer().AddCard(deCard(4))
	bj.GetDealer().AddCard(deCard(5))
	bj.SetPhase(BJPhaseAction)

	assert.Equal(t, BJSuggestStand, GetBasicStrategyAction(hand, deCard(4), false))
	assert.Equal(t, BJSuggestHit, bj.GetBasicStrategySuggestion())
}

func TestStandardGame_DoesNotExposeDealerHoleToStrategy(t *testing.T) {
	bj := NewDefaultBlackJack()
	bj.ToggleHint()
	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	hand.AddCard(deCard(9))
	hand.AddCard(deCard(3))
	bj.GetDealer().AddCard(deCard(4))
	bj.GetDealer().AddCard(deCard(5))
	bj.SetPhase(BJPhaseAction)

	assert.Equal(t, BJSuggestStand, bj.GetBasicStrategySuggestion())
}
