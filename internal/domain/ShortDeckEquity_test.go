package domain_test

import (
	"math/rand"
	"testing"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestCalcShortDeckEquity(t *testing.T) {
	t.Run("pocket aces vs 1 opponent preflop", func(t *testing.T) {
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := domain.CalcShortDeckEquity(humanCards, nil, 1, 5000, rng)
		// Pocket Aces are very strong in ShortDeck
		assert.Greater(t, result.Equity, 0.50)
		assert.Less(t, result.Equity, 0.99)
	})

	t.Run("river deterministic result", func(t *testing.T) {
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
		}
		communityCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 6, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignHeart, 8, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := domain.CalcShortDeckEquity(humanCards, communityCards, 1, 5000, rng)
		// AA on a board with a straight possible - still reasonable equity
		assert.Greater(t, result.Equity, 0.0)
	})

	t.Run("0 opponents returns equity 1.0", func(t *testing.T) {
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := domain.CalcShortDeckEquity(humanCards, nil, 0, 5000, rng)
		assert.Equal(t, 1.0, result.Equity)
	})

	t.Run("0 simulations returns equity 0.0", func(t *testing.T) {
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := domain.CalcShortDeckEquity(humanCards, nil, 1, 0, rng)
		assert.Equal(t, 0.0, result.Equity)
	})

	t.Run("HandOdds probabilities sum to approximately 1.0", func(t *testing.T) {
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
		}
		communityCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 6, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := domain.CalcShortDeckEquity(humanCards, communityCards, 1, 5000, rng)
		sum := 0.0
		for _, h := range result.HandOdds {
			sum += h.Probability
		}
		assert.InDelta(t, 1.0, sum, 0.01)
	})

	t.Run("nil rng uses global rand", func(t *testing.T) {
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
		}
		result := domain.CalcShortDeckEquity(humanCards, nil, 1, 100, nil)
		assert.Greater(t, result.Equity, 0.0)
	})
}

func TestCalcShortDeckEquity_HandOdds(t *testing.T) {
	t.Run("HandOdds entries have correct hand names", func(t *testing.T) {
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
		}
		communityCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 6, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := domain.CalcShortDeckEquity(humanCards, communityCards, 1, 1000, rng)

		assert.Len(t, result.HandOdds, len(domain.ShortDeckHandNames))
		for i, ho := range result.HandOdds {
			assert.Equal(t, i, ho.HandRank)
			assert.Equal(t, domain.ShortDeckHandNames[i], ho.HandName)
			assert.GreaterOrEqual(t, ho.Probability, 0.0)
			assert.LessOrEqual(t, ho.Probability, 1.0)
		}
	})
}

func TestCalcShortDeckEquity_NeededCardsExceedsPool(t *testing.T) {
	t.Run("too many opponents for available pool", func(t *testing.T) {
		// 2 human cards + 5 community = 7 known, pool = 36-7=29
		// With 15 opponents, neededCards = 0 + 15*2 = 30 > 29
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
		}
		communityCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 6, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignHeart, 8, false),
			domain.NewCard(domain.CardDesignClover, 10, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := domain.CalcShortDeckEquity(humanCards, communityCards, 15, 100, rng)
		// All simulations skipped -> wins=0 -> equity=0
		assert.Equal(t, 0.0, result.Equity)
	})
}

func TestCalcShortDeckEquity_RiverExact(t *testing.T) {
	t.Run("river with unbeatable hand has equity 1.0", func(t *testing.T) {
		// Royal flush: A-K spade in hole + Q-J-T spade on community
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignSpade, 13, false),
		}
		communityCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 12, false),
			domain.NewCard(domain.CardDesignSpade, 11, false),
			domain.NewCard(domain.CardDesignSpade, 10, false),
			domain.NewCard(domain.CardDesignHeart, 6, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := domain.CalcShortDeckEquity(humanCards, communityCards, 1, 5000, rng)
		assert.Equal(t, 1.0, result.Equity)
	})
}

func TestCalcShortDeckEquity_ParallelResultsInExpectedRange(t *testing.T) {
	t.Run("parallel execution produces statistically valid results", func(t *testing.T) {
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
		}
		communityCards := []*domain.Card{
			domain.NewCard(domain.CardDesignClover, 6, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := domain.CalcShortDeckEquity(humanCards, communityCards, 1, 50000, rng)
		// AA on a low flop should have reasonable equity
		assert.Greater(t, result.Equity, 0.30)
		assert.Less(t, result.Equity, 0.99)

		// HandOdds probabilities should sum to ~1.0
		sum := 0.0
		for _, h := range result.HandOdds {
			sum += h.Probability
		}
		assert.InDelta(t, 1.0, sum, 0.01)
	})
}

func TestCalcShortDeckEquity_DeterministicWithSeededRng(t *testing.T) {
	t.Run("same seed produces same result", func(t *testing.T) {
		humanCards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 1, false),
		}
		rng1 := rand.New(rand.NewSource(123))
		result1 := domain.CalcShortDeckEquity(humanCards, nil, 1, 1000, rng1)

		rng2 := rand.New(rand.NewSource(123))
		result2 := domain.CalcShortDeckEquity(humanCards, nil, 1, 1000, rng2)

		assert.Equal(t, result1.Equity, result2.Equity)
		for i := range result1.HandOdds {
			assert.Equal(t, result1.HandOdds[i].Probability, result2.HandOdds[i].Probability)
		}
	})
}

func BenchmarkCalcShortDeckEquity(b *testing.B) {
	humanCards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
	}
	communityCards := []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 6, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		domain.CalcShortDeckEquity(humanCards, communityCards, 1, 50000, rng)
	}
}
