package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcOmahaEquity(t *testing.T) {
	t.Run("pocket aces with suited connectors vs 1 opponent preflop", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignHeart, 13, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaEquity(humanCards, nil, 1, 5000, rng)
		// Double suited AA-KK is very strong in Omaha
		assert.Greater(t, result.Equity, 0.50)
		assert.Less(t, result.Equity, 0.95)
	})

	t.Run("river deterministic result", func(t *testing.T) {
		// AA-KK double paired → pair of aces + pair of kings = two pair on low board
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 6, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaEquity(humanCards, communityCards, 1, 5000, rng)
		// AA-KK two pair on low board — in Omaha equity runs closer than Holdem
		assert.Greater(t, result.Equity, 0.30)
	})

	t.Run("0 opponents returns equity 1.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaEquity(humanCards, nil, 0, 5000, rng)
		assert.Equal(t, 1.0, result.Equity)
	})

	t.Run("0 simulations returns equity 0.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaEquity(humanCards, nil, 1, 0, rng)
		assert.Equal(t, 0.0, result.Equity)
	})

	t.Run("HandOdds probabilities sum to approximately 1.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaEquity(humanCards, communityCards, 1, 5000, rng)
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
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
		}
		result := CalcOmahaEquity(humanCards, nil, 1, 100, nil)
		assert.Greater(t, result.Equity, 0.0)
	})
}

func TestCalcOmahaEquity_HandOdds(t *testing.T) {
	t.Run("HandOdds entries have correct hand names", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaEquity(humanCards, communityCards, 1, 1000, rng)

		assert.Len(t, result.HandOdds, len(PokerHandNames))
		for i, ho := range result.HandOdds {
			assert.Equal(t, i, ho.HandRank)
			assert.Equal(t, PokerHandNames[i], ho.HandName)
			assert.GreaterOrEqual(t, ho.Probability, 0.0)
			assert.LessOrEqual(t, ho.Probability, 1.0)
		}
	})
}

func TestCalcOmahaEquity_NeededCardsExceedsPool(t *testing.T) {
	t.Run("too many opponents for available pool", func(t *testing.T) {
		// 4 human cards + 5 community = 9 known, pool = 43
		// With 12 opponents, neededCards = 0 + 12*4 = 48 > 43
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 6, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaEquity(humanCards, communityCards, 12, 100, rng)
		// All simulations skipped → wins=0 → equity=0
		assert.Equal(t, 0.0, result.Equity)
	})
}

func TestCalcOmahaEquity_RiverExact(t *testing.T) {
	t.Run("river with unbeatable hand has equity 1.0", func(t *testing.T) {
		// Royal flush: A♠ K♠ in hole + Q♠ J♠ T♠ on community
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignHeart, 2, false),
			NewCard(CardDesignClover, 3, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignSpade, 12, false),
			NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 5, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaEquity(humanCards, communityCards, 1, 5000, rng)
		assert.Equal(t, 1.0, result.Equity)
	})
}

func TestCalcOmahaEquity_ParallelResultsInExpectedRange(t *testing.T) {
	t.Run("parallel execution produces statistically valid results", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignHeart, 13, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaEquity(humanCards, communityCards, 1, 50000, rng)
		// AA-KK double suited on a low flop should have reasonable equity
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

func TestCalcOmahaEquity_DeterministicWithSeededRng(t *testing.T) {
	t.Run("same seed produces same result", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
		}
		rng1 := rand.New(rand.NewSource(123))
		result1 := CalcOmahaEquity(humanCards, nil, 1, 1000, rng1)

		rng2 := rand.New(rand.NewSource(123))
		result2 := CalcOmahaEquity(humanCards, nil, 1, 1000, rng2)

		assert.Equal(t, result1.Equity, result2.Equity)
		for i := range result1.HandOdds {
			assert.Equal(t, result1.HandOdds[i].Probability, result2.HandOdds[i].Probability)
		}
	})
}

func TestCalcOmahaHiLoEquity(t *testing.T) {
	t.Run("no qualifying low possible produces identical equity to Hi-only and zero LowProbability", func(t *testing.T) {
		// Quad Aces on a high board: human holds 2 Aces, board has 2 Aces and three cards >= 9.
		// No low can qualify (only 1 distinct rank <= 8 on board; Omaha low requires 3 distinct ranks <= 8 from board).
		// Opponents cannot beat or tie Quad Aces.
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignHeart, 13, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignDiamond, 1, false),
			NewCard(CardDesignClover, 1, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignDiamond, 10, false),
			NewCard(CardDesignClover, 11, false),
		}
		rng1 := rand.New(rand.NewSource(42))
		rng2 := rand.New(rand.NewSource(42))

		hiResult := CalcOmahaEquity(humanCards, communityCards, 1, 500, rng1)
		hiloResult := CalcOmahaHiLoEquity(humanCards, communityCards, 1, 500, rng2)

		assert.Equal(t, 1.0, hiResult.Equity)
		assert.Equal(t, hiResult.Equity, hiloResult.Equity)
		assert.Equal(t, 0.0, hiloResult.LowProbability)
	})

	t.Run("0 opponents returns equity 1.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 2, false),
			NewCard(CardDesignClover, 3, false),
			NewCard(CardDesignDiamond, 4, false),
		}
		result := CalcOmahaHiLoEquity(humanCards, nil, 0, 500, nil)
		assert.Equal(t, 1.0, result.Equity)
		assert.Len(t, result.HandOdds, len(PokerHandNames))
	})

	t.Run("0 simulations returns zero equity", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 2, false),
			NewCard(CardDesignClover, 3, false),
			NewCard(CardDesignDiamond, 4, false),
		}
		result := CalcOmahaHiLoEquity(humanCards, nil, 1, 0, nil)
		assert.Equal(t, 0.0, result.Equity)
		assert.Len(t, result.HandOdds, len(PokerHandNames))
	})

	t.Run("low-only winning hand has positive share (> 0) despite losing high", func(t *testing.T) {
		// Board: 3c, 4c, 8c, Ks, Qs (three low cards: 3, 4, 8)
		// Human has: Ah, 2h, 6d, 7d -> nut low (A-2-3-4-8), high is at best King-high
		// Even if high is lost, human's share of the split pot is > 0
		humanCards := []*Card{
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignHeart, 2, false),
			NewCard(CardDesignDiamond, 6, false),
			NewCard(CardDesignDiamond, 7, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 3, false),
			NewCard(CardDesignClover, 4, false),
			NewCard(CardDesignClover, 8, false),
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignSpade, 12, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaHiLoEquity(humanCards, communityCards, 1, 2000, rng)

		// Low half share: nut low wins almost every time (or ties with another A-2)
		assert.Greater(t, result.LowProbability, 0.60)
		assert.LessOrEqual(t, result.LowProbability, 1.0)

		// Pot share: since 0.5 * lowShare > 0, overall equity is strictly positive
		assert.Greater(t, result.Equity, 0.30)
		assert.Less(t, result.Equity, 0.65)
	})

	t.Run("scoop hand approaches 1.0 equity", func(t *testing.T) {
		// Community: 3d, 4d, 5d, Kc, Qc
		// Human: Ad, 2d, As, 2s -> Straight Flush (A-2-3-4-5 diamonds) + Wheel nut low (A-2-3-4-5)
		// Scoops both high and low pots
		humanCards := []*Card{
			NewCard(CardDesignDiamond, 1, false),
			NewCard(CardDesignDiamond, 2, false),
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 2, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignDiamond, 3, false),
			NewCard(CardDesignDiamond, 4, false),
			NewCard(CardDesignDiamond, 5, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignClover, 12, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaHiLoEquity(humanCards, communityCards, 1, 2000, rng)

		assert.Greater(t, result.Equity, 0.95)
		assert.LessOrEqual(t, result.Equity, 1.0)
		assert.Greater(t, result.LowProbability, 0.95)
		assert.LessOrEqual(t, result.LowProbability, 1.0)
	})

	t.Run("too many opponents for available deck returns zero", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 2, false),
			NewCard(CardDesignClover, 3, false),
			NewCard(CardDesignDiamond, 4, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaHiLoEquity(humanCards, communityCards, 12, 100, rng)
		assert.Equal(t, 0.0, result.Equity)
	})

	t.Run("HandOdds probabilities sum to approximately 1.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 2, false),
			NewCard(CardDesignClover, 3, false),
			NewCard(CardDesignDiamond, 4, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcOmahaHiLoEquity(humanCards, communityCards, 1, 1000, rng)

		sum := 0.0
		for _, h := range result.HandOdds {
			sum += h.Probability
		}
		assert.InDelta(t, 1.0, sum, 0.01)
	})
}

func BenchmarkCalcOmahaEquity(b *testing.B) {
	humanCards := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignDiamond, 13, false),
	}
	communityCards := []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignSpade, 9, false),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		CalcOmahaEquity(humanCards, communityCards, 1, 50000, rng)
	}
}
