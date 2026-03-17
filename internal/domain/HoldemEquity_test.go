package domain

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcEquity(t *testing.T) {
	t.Run("pocket aces vs 1 opponent preflop", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcEquity(humanCards, nil, 1, 5000, rng)
		// Pocket aces vs 1 opponent: ~80-90% equity
		assert.Greater(t, result.Equity, 0.70)
		assert.Less(t, result.Equity, 0.95)
	})

	t.Run("river deterministic result", func(t *testing.T) {
		// Human has pair of aces, opponent gets random cards but community is fully dealt
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 6, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcEquity(humanCards, communityCards, 1, 5000, rng)
		// With pair of aces on a low board, equity should be very high
		assert.Greater(t, result.Equity, 0.80)
	})

	t.Run("0 opponents returns equity 1.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcEquity(humanCards, nil, 0, 5000, rng)
		assert.Equal(t, 1.0, result.Equity)
	})

	t.Run("0 simulations returns equity 0.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcEquity(humanCards, nil, 1, 0, rng)
		assert.Equal(t, 0.0, result.Equity)
	})

	t.Run("HandOdds probabilities sum to approximately 1.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcEquity(humanCards, communityCards, 1, 5000, rng)
		sum := 0.0
		for _, h := range result.HandOdds {
			sum += h.Probability
		}
		assert.InDelta(t, 1.0, sum, 0.01)
	})

	t.Run("nil rng uses global rand", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
		}
		result := CalcEquity(humanCards, nil, 1, 100, nil)
		assert.Greater(t, result.Equity, 0.0)
	})
}

func TestCalcPotOdds(t *testing.T) {
	t.Run("pot=100 call=50 returns 33.33", func(t *testing.T) {
		result := CalcPotOdds(100, 50)
		assert.InDelta(t, 33.33, result, 0.01)
	})

	t.Run("call=0 returns 0.0", func(t *testing.T) {
		result := CalcPotOdds(100, 0)
		assert.Equal(t, 0.0, result)
	})

	t.Run("pot=0 call=10 returns 100.0", func(t *testing.T) {
		result := CalcPotOdds(0, 10)
		assert.Equal(t, 100.0, result)
	})

	t.Run("pot=200 call=100 returns 33.33", func(t *testing.T) {
		result := CalcPotOdds(200, 100)
		expected := float64(100) / float64(200+100) * 100
		assert.InDelta(t, expected, result, 0.01)
	})
}

func TestCalcEquity_HandOdds(t *testing.T) {
	t.Run("HandOdds entries have correct hand names", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcEquity(humanCards, communityCards, 1, 1000, rng)

		assert.Len(t, result.HandOdds, len(PokerHandNames))
		for i, ho := range result.HandOdds {
			assert.Equal(t, i, ho.HandRank)
			assert.Equal(t, PokerHandNames[i], ho.HandName)
			assert.GreaterOrEqual(t, ho.Probability, 0.0)
			assert.LessOrEqual(t, ho.Probability, 1.0)
		}
	})
}

func TestCalcEquity_RiverExact(t *testing.T) {
	t.Run("river with unbeatable hand has equity 1.0", func(t *testing.T) {
		// Royal flush: A K Q J 10 of spades
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 13, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignSpade, 12, false),
			NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignHeart, 2, false),
			NewCard(CardDesignClover, 3, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcEquity(humanCards, communityCards, 1, 5000, rng)
		assert.Equal(t, 1.0, result.Equity)
	})
}

func TestCalcPotOdds_NegativeCallAmount(t *testing.T) {
	// Negative call shouldn't happen but ensure no panic
	result := CalcPotOdds(100, -10)
	assert.True(t, math.IsInf(result, 0) || result < 0 || result >= 0) // just no panic
}
