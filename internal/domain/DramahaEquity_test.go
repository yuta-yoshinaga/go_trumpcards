package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcDramahaEquity(t *testing.T) {
	t.Run("pocket aces with suited connectors vs 1 opponent preflop", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignHeart, 13, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcDramahaEquity(humanCards, nil, 1, 5000, rng)
		// Double suited AA-KK is very strong in Dramaha
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
			// ドラマハは常に 5 枚。4 枚の手札で相手に 5 枚配ると、卓では
			// 起こりえない不利な比較になる。
			NewCard(CardDesignSpade, 12, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 6, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcDramahaEquity(humanCards, communityCards, 1, 5000, rng)
		// **AA-KK-Q は 2-7-9-4-6 のレインボー盤で「エースのワンペア」止まり。**
		// オマハ式に手札ちょうど 2 枚を使うので、ポケットペアは盤が塗り替わらない
		// 限り 1 ペアにしかならない (元のコメントの "two pair" は誤り)。
		// 相手は 5 枚から C(5,2)=10 通りを持つため、裸のオーバーペアは後手に回る。
		// 閾値 0.30 は相手に 4 枚しか配っていなかった頃の数字。
		assert.Greater(t, result.Equity, 0.10, "0 や壊れた値を返していないこと")
		assert.Less(t, result.Equity, 0.35, "裸のワンペアがこの盤で優勢になることはない")
	})

	t.Run("0 opponents returns equity 1.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
			// ドラマハは常に 5 枚。4 枚の手札で相手に 5 枚配ると、卓では
			// 起こりえない不利な比較になる。
			NewCard(CardDesignSpade, 12, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcDramahaEquity(humanCards, nil, 0, 5000, rng)
		assert.Equal(t, 1.0, result.Equity)
	})

	t.Run("0 simulations returns equity 0.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
			// ドラマハは常に 5 枚。4 枚の手札で相手に 5 枚配ると、卓では
			// 起こりえない不利な比較になる。
			NewCard(CardDesignSpade, 12, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcDramahaEquity(humanCards, nil, 1, 0, rng)
		assert.Equal(t, 0.0, result.Equity)
	})

	t.Run("HandOdds probabilities sum to approximately 1.0", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
			// ドラマハは常に 5 枚。4 枚の手札で相手に 5 枚配ると、卓では
			// 起こりえない不利な比較になる。
			NewCard(CardDesignSpade, 12, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcDramahaEquity(humanCards, communityCards, 1, 5000, rng)
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
			// ドラマハは常に 5 枚。4 枚の手札で相手に 5 枚配ると、卓では
			// 起こりえない不利な比較になる。
			NewCard(CardDesignSpade, 12, false),
		}
		result := CalcDramahaEquity(humanCards, nil, 1, 100, nil)
		assert.Greater(t, result.Equity, 0.0)
	})
}

func TestCalcDramahaEquity_HandOdds(t *testing.T) {
	t.Run("HandOdds entries have correct hand names", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
			// ドラマハは常に 5 枚。4 枚の手札で相手に 5 枚配ると、卓では
			// 起こりえない不利な比較になる。
			NewCard(CardDesignSpade, 12, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcDramahaEquity(humanCards, communityCards, 1, 1000, rng)

		assert.Len(t, result.HandOdds, len(PokerHandNames))
		for i, ho := range result.HandOdds {
			assert.Equal(t, i, ho.HandRank)
			assert.Equal(t, PokerHandNames[i], ho.HandName)
			assert.GreaterOrEqual(t, ho.Probability, 0.0)
			assert.LessOrEqual(t, ho.Probability, 1.0)
		}
	})
}

func TestCalcDramahaEquity_NeededCardsExceedsPool(t *testing.T) {
	t.Run("too many opponents for available pool", func(t *testing.T) {
		// 4 human cards + 5 community = 9 known, pool = 43
		// With 12 opponents, neededCards = 0 + 12*4 = 48 > 43
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
			// ドラマハは常に 5 枚。4 枚の手札で相手に 5 枚配ると、卓では
			// 起こりえない不利な比較になる。
			NewCard(CardDesignSpade, 12, false),
		}
		communityCards := []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignSpade, 9, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 6, false),
		}
		rng := rand.New(rand.NewSource(42))
		result := CalcDramahaEquity(humanCards, communityCards, 12, 100, rng)
		// All simulations skipped → wins=0 → equity=0
		assert.Equal(t, 0.0, result.Equity)
	})
}

func TestCalcDramahaEquity_RiverExact(t *testing.T) {
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
		result := CalcDramahaEquity(humanCards, communityCards, 1, 5000, rng)
		assert.Equal(t, 1.0, result.Equity)
	})
}

func TestCalcDramahaEquity_ParallelResultsInExpectedRange(t *testing.T) {
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
		result := CalcDramahaEquity(humanCards, communityCards, 1, 50000, rng)
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

func TestCalcDramahaEquity_DeterministicWithSeededRng(t *testing.T) {
	t.Run("same seed produces same result", func(t *testing.T) {
		humanCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignClover, 13, false),
			NewCard(CardDesignDiamond, 13, false),
			// ドラマハは常に 5 枚。4 枚の手札で相手に 5 枚配ると、卓では
			// 起こりえない不利な比較になる。
			NewCard(CardDesignSpade, 12, false),
		}
		rng1 := rand.New(rand.NewSource(123))
		result1 := CalcDramahaEquity(humanCards, nil, 1, 1000, rng1)

		rng2 := rand.New(rand.NewSource(123))
		result2 := CalcDramahaEquity(humanCards, nil, 1, 1000, rng2)

		assert.Equal(t, result1.Equity, result2.Equity)
		for i := range result1.HandOdds {
			assert.Equal(t, result1.HandOdds[i].Probability, result2.HandOdds[i].Probability)
		}
	})
}

func BenchmarkCalcDramahaEquity(b *testing.B) {
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
		CalcDramahaEquity(humanCards, communityCards, 1, 50000, rng)
	}
}
