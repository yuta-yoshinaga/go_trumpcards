package domain

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeOddsPoker(jokerCount int) (*Poker, []*PokerPlayer) {
	tc := NewTrumpCards(jokerCount)
	players := []*PokerPlayer{
		NewPokerPlayer(true, PokerStyleBalanced),
		NewPokerPlayer(false, PokerStyleConservative),
	}
	cfg := DefaultPokerConfig()
	cfg.JokerCount = jokerCount
	cfg.CpuCount = 1
	p := NewPoker(tc, players, cfg)
	return p, players
}

func TestCalcDrawOdds_WrongPhase(t *testing.T) {
	p, _ := makeOddsPoker(0)

	// Phase: Init
	p.SetPhase(PokerPhaseInit)
	p.SetCurrentTurn(0)
	_, err := p.CalcDrawOdds([]int{0})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)

	// Phase: Deal
	p.SetPhase(PokerPhaseDeal)
	_, err = p.CalcDrawOdds([]int{0})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)

	// Phase: SecondBet
	p.SetPhase(PokerPhaseSecondBet)
	_, err = p.CalcDrawOdds([]int{0})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)

	// Phase: End
	p.SetPhase(PokerPhaseEnd)
	_, err = p.CalcDrawOdds([]int{0})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestCalcDrawOdds_NotHumanTurn(t *testing.T) {
	p, _ := makeOddsPoker(0)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(1) // CPU player

	_, err := p.CalcDrawOdds([]int{0})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestCalcDrawOdds_InvalidIndices_OutOfRange(t *testing.T) {
	p, players := makeOddsPoker(0)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	// Give human 5 cards
	for i := 1; i <= 5; i++ {
		players[0].AddCard(NewCard(CardDesignSpade, i, false))
	}

	_, err := p.CalcDrawOdds([]int{5})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidIndices)

	_, err = p.CalcDrawOdds([]int{-1})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidIndices)
}

func TestCalcDrawOdds_InvalidIndices_Duplicate(t *testing.T) {
	p, players := makeOddsPoker(0)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	for i := 1; i <= 5; i++ {
		players[0].AddCard(NewCard(CardDesignSpade, i, false))
	}

	_, err := p.CalcDrawOdds([]int{0, 0})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidIndices)
}

func TestCalcDrawOdds_StandZeroIndices(t *testing.T) {
	p, players := makeOddsPoker(0)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	// One Pair (two 10s)
	players[0].AddCard(NewCard(CardDesignSpade, 10, false))
	players[0].AddCard(NewCard(CardDesignHeart, 10, false))
	players[0].AddCard(NewCard(CardDesignClover, 3, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 5, false))
	players[0].AddCard(NewCard(CardDesignSpade, 7, false))

	odds, err := p.CalcDrawOdds([]int{})
	assert.NoError(t, err)
	assert.Len(t, odds, len(PokerHandNames))

	// Current hand is One Pair → 100%
	assert.Equal(t, 1.0, odds[PokerHandOnePair].Probability)
	assert.Equal(t, 1, odds[PokerHandOnePair].Count)
	assert.Equal(t, 1, odds[PokerHandOnePair].Total)

	// All others 0%
	for i, o := range odds {
		if i != PokerHandOnePair {
			assert.Equal(t, 0.0, o.Probability)
			assert.Equal(t, 0, o.Count)
		}
	}
}

func TestCalcDrawOdds_FlushDraw(t *testing.T) {
	// 4 spades + 1 heart, exchange the heart → verify flush probability
	p, players := makeOddsPoker(0)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	// 4 spades: A, K, Q, J + 1 heart: 2
	players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	players[0].AddCard(NewCard(CardDesignSpade, 13, false))
	players[0].AddCard(NewCard(CardDesignSpade, 12, false))
	players[0].AddCard(NewCard(CardDesignSpade, 11, false))
	players[0].AddCard(NewCard(CardDesignHeart, 2, false))

	odds, err := p.CalcDrawOdds([]int{4}) // exchange the heart 2
	assert.NoError(t, err)

	// Pool = 52 - 5 = 47 cards
	assert.Equal(t, 47, odds[0].Total)

	// Flush probability: 9 remaining spades out of 47 cards
	// But some of those spades will make Straight Flush or Royal Flush
	// Spade 10 → Royal Flush (A,K,Q,J,10)
	// So: Royal Flush = 1, Flush = 8, Straight = some, others remaining
	// Actually: A,K,Q,J + any spade = at least Flush
	// Spade 10 → A,K,Q,J,10 all spades → Royal Flush
	// So Royal Flush count = 1, Straight Flush = 0 (10 already counted as Royal)
	// Flush count = 9 - 1 = 8 (excluding Royal Flush)
	// Note: need to account for Straight as well
	// Non-spade 10 → A,K,Q,J,10 → Straight (not same suit)
	// Let me just check the sum = 47 and flush-related probs
	assert.Equal(t, 1, odds[PokerHandRoyalFlush].Count) // Spade 10

	// 9 spades total: 9 of them make flush or better
	flushOrBetter := odds[PokerHandFlush].Count + odds[PokerHandStraightFlush].Count + odds[PokerHandRoyalFlush].Count
	assert.Equal(t, 9, flushOrBetter)

	// Probabilities sum to 1.0
	var sumProb float64
	sumCount := 0
	for _, o := range odds {
		sumProb += o.Probability
		sumCount += o.Count
	}
	assert.InDelta(t, 1.0, sumProb, 1e-9)
	assert.Equal(t, 47, sumCount)
}

func TestCalcDrawOdds_WithJoker1(t *testing.T) {
	// With 1 joker in config, pool should have 47 cards (52 + 1 - 5 - 1 if human has joker)
	// or 48 cards (52 + 1 - 5) if human doesn't have joker
	p, players := makeOddsPoker(1)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	// Human has no joker
	players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	players[0].AddCard(NewCard(CardDesignHeart, 2, false))
	players[0].AddCard(NewCard(CardDesignClover, 3, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 4, false))
	players[0].AddCard(NewCard(CardDesignSpade, 6, false))

	odds, err := p.CalcDrawOdds([]int{4}) // exchange 1 card
	assert.NoError(t, err)
	// Pool = 52 - 5 + 1 joker = 48
	assert.Equal(t, 48, odds[0].Total)
}

func TestCalcDrawOdds_WithJoker2(t *testing.T) {
	// 2 jokers, human has 1 joker
	p, players := makeOddsPoker(2)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	players[0].AddCard(NewCard(CardDesignJoker, CardValueJoker, false))
	players[0].AddCard(NewCard(CardDesignHeart, 2, false))
	players[0].AddCard(NewCard(CardDesignClover, 3, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 4, false))
	players[0].AddCard(NewCard(CardDesignSpade, 6, false))

	odds, err := p.CalcDrawOdds([]int{4}) // exchange 1 card
	assert.NoError(t, err)
	// Pool = 52 - 4 non-joker human cards + (2 - 1) remaining jokers = 48 + 1 = 49
	assert.Equal(t, 49, odds[0].Total)
}

func TestCalcDrawOdds_ProbsSumToOne(t *testing.T) {
	p, players := makeOddsPoker(0)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	players[0].AddCard(NewCard(CardDesignHeart, 5, false))
	players[0].AddCard(NewCard(CardDesignClover, 9, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 11, false))
	players[0].AddCard(NewCard(CardDesignSpade, 3, false))

	// Exchange 3 cards
	odds, err := p.CalcDrawOdds([]int{1, 2, 4})
	assert.NoError(t, err)

	var sumProb float64
	sumCount := 0
	for _, o := range odds {
		sumProb += o.Probability
		sumCount += o.Count
	}
	assert.InDelta(t, 1.0, sumProb, 1e-9)
	assert.Equal(t, odds[0].Total, sumCount)
}

func TestCalcDrawOdds_ExchangeAll5(t *testing.T) {
	p, players := makeOddsPoker(0)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	players[0].AddCard(NewCard(CardDesignHeart, 5, false))
	players[0].AddCard(NewCard(CardDesignClover, 9, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 11, false))
	players[0].AddCard(NewCard(CardDesignSpade, 3, false))

	odds, err := p.CalcDrawOdds([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)

	// C(47, 5) = 1533939
	assert.Equal(t, 1533939, odds[0].Total)

	var sumProb float64
	sumCount := 0
	for _, o := range odds {
		sumProb += o.Probability
		sumCount += o.Count
	}
	assert.InDelta(t, 1.0, sumProb, 1e-9)
	assert.Equal(t, 1533939, sumCount)
}

func TestCalcDrawOdds_KnownFlushProbability(t *testing.T) {
	// More specific: 4 spades (2,3,4,5) + 1 heart 9, exchange heart 9
	// Remaining spades: 1,6,7,8,9,10,11,12,13 = 9 spades
	// But some make straight flush: Spade 6 → 2,3,4,5,6 straight flush
	// Spade 1 → A,2,3,4,5 straight flush
	// So: straight flush = 2, flush = 7
	p, players := makeOddsPoker(0)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	players[0].AddCard(NewCard(CardDesignSpade, 2, false))
	players[0].AddCard(NewCard(CardDesignSpade, 3, false))
	players[0].AddCard(NewCard(CardDesignSpade, 4, false))
	players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	players[0].AddCard(NewCard(CardDesignHeart, 9, false))

	odds, err := p.CalcDrawOdds([]int{4})
	assert.NoError(t, err)

	// Total flush or better = 9/47
	flushOrBetter := odds[PokerHandFlush].Count + odds[PokerHandStraightFlush].Count + odds[PokerHandRoyalFlush].Count
	assert.Equal(t, 9, flushOrBetter)
	assert.InDelta(t, 9.0/47.0, float64(flushOrBetter)/float64(odds[0].Total), 1e-9)
}

// --- Helper function tests ---

func TestBuildUnknownCards_NoJoker(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
	}
	pool := buildUnknownCards(hand, 0)
	assert.Equal(t, 47, len(pool))
}

func TestBuildUnknownCards_WithJoker_HumanHasNone(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
	}
	pool := buildUnknownCards(hand, 2)
	// 52 - 5 + 2 jokers = 49
	assert.Equal(t, 49, len(pool))
}

func TestBuildUnknownCards_WithJoker_HumanHasOne(t *testing.T) {
	hand := []*Card{
		NewCard(CardDesignJoker, CardValueJoker, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 5, false),
	}
	pool := buildUnknownCards(hand, 2)
	// 52 - 4 + (2 - 1) jokers = 49
	assert.Equal(t, 49, len(pool))
}

func TestPokerCombinations_Empty(t *testing.T) {
	pool := []*Card{NewCard(CardDesignSpade, 1, false)}
	count := 0
	pokerCombinations(pool, 0, func(combo []*Card) {
		count++
	})
	assert.Equal(t, 0, count)
}

func TestPokerCombinations_KGreaterThanN(t *testing.T) {
	pool := []*Card{NewCard(CardDesignSpade, 1, false)}
	count := 0
	pokerCombinations(pool, 2, func(combo []*Card) {
		count++
	})
	assert.Equal(t, 0, count)
}

func TestPokerCombinations_C5_2(t *testing.T) {
	pool := make([]*Card, 5)
	for i := 0; i < 5; i++ {
		pool[i] = NewCard(CardDesignSpade, i+1, false)
	}
	count := 0
	pokerCombinations(pool, 2, func(combo []*Card) {
		count++
	})
	// C(5,2) = 10
	assert.Equal(t, 10, count)
}

func TestPokerCombinations_C47_1(t *testing.T) {
	pool := make([]*Card, 47)
	for i := 0; i < 47; i++ {
		pool[i] = NewCard(CardDesignSpade, (i%13)+1, false)
	}
	count := 0
	pokerCombinations(pool, 1, func(combo []*Card) {
		count++
	})
	assert.Equal(t, 47, count)
}

// Verify C(n,k) formula
func comb(n, k int) int {
	if k > n {
		return 0
	}
	result := 1
	for i := 0; i < k; i++ {
		result = result * (n - i) / (i + 1)
	}
	return result
}

func TestPokerCombinations_VariousK(t *testing.T) {
	pool := make([]*Card, 10)
	for i := 0; i < 10; i++ {
		pool[i] = NewCard(CardDesignSpade, (i%13)+1, false)
	}
	for k := 1; k <= 5; k++ {
		count := 0
		pokerCombinations(pool, k, func(combo []*Card) {
			count++
		})
		expected := comb(10, k)
		assert.Equal(t, expected, count, "C(10,%d)", k)
	}
}

func TestCalcDrawOdds_SingleExchangeExactProbability(t *testing.T) {
	// Simple case: pair of aces + 3 others, exchange 1 card
	// We can verify specific properties
	p, players := makeOddsPoker(0)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	players[0].AddCard(NewCard(CardDesignHeart, 1, false))
	players[0].AddCard(NewCard(CardDesignClover, 3, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 7, false))
	players[0].AddCard(NewCard(CardDesignSpade, 9, false))

	odds, err := p.CalcDrawOdds([]int{4}) // exchange the 9
	assert.NoError(t, err)
	assert.Equal(t, 47, odds[0].Total)

	// Three of a Kind: need another Ace → 2 remaining aces
	assert.Equal(t, 2, odds[PokerHandThreeOfAKind].Count)
	assert.InDelta(t, 2.0/47.0, odds[PokerHandThreeOfAKind].Probability, 1e-9)

	// Two Pair: need another 3 (3 cards) or another 7 (3 cards) = 6 total
	assert.Equal(t, 6, odds[PokerHandTwoPair].Count)

	// All probs sum to 1
	var sum float64
	for _, o := range odds {
		sum += o.Probability
	}
	assert.InDelta(t, 1.0, sum, 1e-9)
}

func TestBuildOddsResult(t *testing.T) {
	result := buildOddsResult(PokerHandFlush, 1)
	assert.Len(t, result, len(PokerHandNames))
	assert.Equal(t, 1.0, result[PokerHandFlush].Probability)
	assert.Equal(t, 1, result[PokerHandFlush].Count)
	assert.Equal(t, 1, result[PokerHandFlush].Total)
	assert.Equal(t, 0.0, result[PokerHandHighCard].Probability)
	assert.Equal(t, "Flush", result[PokerHandFlush].HandName)
}

func TestCalcDrawOdds_FiveOfAKindWithJokers(t *testing.T) {
	// 2 jokers, human has 4 aces + 1 card to exchange
	// Exchanging a non-ace should show FiveOfAKind possible
	p, players := makeOddsPoker(2)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	players[0].AddCard(NewCard(CardDesignHeart, 1, false))
	players[0].AddCard(NewCard(CardDesignClover, 1, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 1, false))
	players[0].AddCard(NewCard(CardDesignSpade, 2, false)) // exchange this

	odds, err := p.CalcDrawOdds([]int{4})
	assert.NoError(t, err)

	// FiveOfAKind is possible when drawing a joker
	// With 2 jokers and human has 0 → 2 jokers in pool
	assert.Equal(t, 2, odds[PokerHandFiveOfAKind].Count)
	assert.True(t, odds[PokerHandFiveOfAKind].Probability > 0)
}

func TestCalcDrawOdds_LargeExchange_C49_5(t *testing.T) {
	// With 2 jokers, exchange all 5 → C(49, 5) = 1906884
	p, players := makeOddsPoker(2)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	players[0].AddCard(NewCard(CardDesignHeart, 2, false))
	players[0].AddCard(NewCard(CardDesignClover, 3, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 4, false))
	players[0].AddCard(NewCard(CardDesignSpade, 6, false))

	odds, err := p.CalcDrawOdds([]int{0, 1, 2, 3, 4})
	assert.NoError(t, err)

	expected := comb(49, 5) // 1906884
	assert.Equal(t, expected, odds[0].Total)

	var sum float64
	for _, o := range odds {
		sum += o.Probability
	}
	assert.InDelta(t, 1.0, sum, 1e-9)
}

func TestCalcDrawOdds_FlushExactProbability_9over47(t *testing.T) {
	// 4 spades (2,4,6,8) + 1 heart K, exchange the heart K
	// 9 remaining spades, some may form straight+flush combos
	p, players := makeOddsPoker(0)
	p.SetPhase(PokerPhaseExchange)
	p.SetCurrentTurn(0)
	players[0].AddCard(NewCard(CardDesignSpade, 2, false))
	players[0].AddCard(NewCard(CardDesignSpade, 4, false))
	players[0].AddCard(NewCard(CardDesignSpade, 6, false))
	players[0].AddCard(NewCard(CardDesignSpade, 8, false))
	players[0].AddCard(NewCard(CardDesignHeart, 13, false))

	odds, err := p.CalcDrawOdds([]int{4})
	assert.NoError(t, err)

	// All 9 remaining spades make flush (no straight flush possible with 2,4,6,8,X)
	flushTotal := odds[PokerHandFlush].Count + odds[PokerHandStraightFlush].Count + odds[PokerHandRoyalFlush].Count
	assert.Equal(t, 9, flushTotal)

	// With 2,4,6,8 gaps: no 5-card sequence possible → all 9 are plain flush
	assert.Equal(t, 9, odds[PokerHandFlush].Count)
	assert.Equal(t, 0, odds[PokerHandStraightFlush].Count)
	assert.Equal(t, 0, odds[PokerHandRoyalFlush].Count)

	// Exact probability 9/47
	assert.True(t, math.Abs(odds[PokerHandFlush].Probability-9.0/47.0) < 1e-9)
}
