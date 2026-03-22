package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBettingHandBracket(t *testing.T) {
	tests := []struct {
		name     string
		handRank int
		want     int
	}{
		{"HighCard → weak(0)", PokerHandHighCard, 0},
		{"OnePair → medium(1)", PokerHandOnePair, 1},
		{"TwoPair → medium(1)", PokerHandTwoPair, 1},
		{"ThreeOfAKind → strong(2)", PokerHandThreeOfAKind, 2},
		{"Straight → strong(2)", PokerHandStraight, 2},
		{"Flush → strong(2)", PokerHandFlush, 2},
		{"FullHouse → strong(2)", PokerHandFullHouse, 2},
		{"FourOfAKind → strong(2)", PokerHandFourOfAKind, 2},
		{"StraightFlush → strong(2)", PokerHandStraightFlush, 2},
		{"RoyalFlush → strong(2)", PokerHandRoyalFlush, 2},
		{"FiveOfAKind → strong(2)", PokerHandFiveOfAKind, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, bettingHandBracket(tt.handRank))
		})
	}
}

func TestBettingHumanProfile_RecordAction(t *testing.T) {
	p := &BettingHumanProfile{}

	// Bet with weak hand (bluff)
	p.RecordAction(PokerHandHighCard, bettingActionBet)
	assert.Equal(t, 1, p.AggressiveByBracket[0].Aggressive)
	assert.Equal(t, 1, p.AggressiveByBracket[0].Total)

	// Call with weak hand (not aggressive)
	p.RecordAction(PokerHandHighCard, bettingActionCall)
	assert.Equal(t, 1, p.AggressiveByBracket[0].Aggressive)
	assert.Equal(t, 2, p.AggressiveByBracket[0].Total)

	// Raise with medium hand
	p.RecordAction(PokerHandOnePair, bettingActionRaise)
	assert.Equal(t, 1, p.AggressiveByBracket[1].Aggressive)
	assert.Equal(t, 1, p.AggressiveByBracket[1].Total)

	// Check with strong hand
	p.RecordAction(PokerHandFullHouse, bettingActionCheck)
	assert.Equal(t, 0, p.AggressiveByBracket[2].Aggressive)
	assert.Equal(t, 1, p.AggressiveByBracket[2].Total)

	// Fold with medium hand (not aggressive)
	p.RecordAction(PokerHandTwoPair, bettingActionFold)
	assert.Equal(t, 1, p.AggressiveByBracket[1].Aggressive)
	assert.Equal(t, 2, p.AggressiveByBracket[1].Total)
}

func TestBettingHumanProfile_RecordFoldToBet(t *testing.T) {
	p := &BettingHumanProfile{}

	p.RecordFoldToBet(true)
	assert.Equal(t, 1, p.FoldToBetCount)
	assert.Equal(t, 1, p.FoldToBetTotal)

	p.RecordFoldToBet(false)
	assert.Equal(t, 1, p.FoldToBetCount)
	assert.Equal(t, 2, p.FoldToBetTotal)
}

func TestBettingHumanProfile_BluffRate(t *testing.T) {
	p := &BettingHumanProfile{}

	// No data → 0.5
	assert.Equal(t, 0.5, p.BluffRate(0))
	assert.Equal(t, 0.5, p.BluffRate(1))
	assert.Equal(t, 0.5, p.BluffRate(2))

	// Out of range → 0.5
	assert.Equal(t, 0.5, p.BluffRate(-1))
	assert.Equal(t, 0.5, p.BluffRate(3))

	// With data
	p.RecordAction(PokerHandHighCard, bettingActionBet)
	p.RecordAction(PokerHandHighCard, bettingActionCall)
	assert.Equal(t, 0.5, p.BluffRate(0))

	p.RecordAction(PokerHandHighCard, bettingActionRaise)
	assert.InDelta(t, 2.0/3.0, p.BluffRate(0), 0.001)
}

func TestBettingHumanProfile_FoldRate(t *testing.T) {
	p := &BettingHumanProfile{}

	// No data → 0.5
	assert.Equal(t, 0.5, p.FoldRate())

	// With data
	p.RecordFoldToBet(true)
	p.RecordFoldToBet(true)
	p.RecordFoldToBet(false)
	assert.InDelta(t, 2.0/3.0, p.FoldRate(), 0.001)
}

func TestBettingHumanProfile_AdaptStrength(t *testing.T) {
	tests := []struct {
		name        string
		gamesPlayed int
		want        float64
	}{
		{"0 games", 0, 0.0},
		{"1 game", 1, 0.04},
		{"3 games", 3, 0.12},
		{"5 games", 5, 0.20},
		{"10 games (capped)", 10, 0.20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &BettingHumanProfile{GamesPlayed: tt.gamesPlayed}
			assert.InDelta(t, tt.want, p.AdaptStrength(), 0.001)
		})
	}
}

func TestBettingHumanProfile_RecordHesitation(t *testing.T) {
	t.Run("ms<=0 is no-op", func(t *testing.T) {
		p := &BettingHumanProfile{}
		p.RecordHesitation(0)
		assert.Equal(t, 0, p.HesitationCount)
		assert.Equal(t, 0.0, p.HesitationMean)

		p.RecordHesitation(-1)
		assert.Equal(t, 0, p.HesitationCount)
	})

	t.Run("single data point", func(t *testing.T) {
		p := &BettingHumanProfile{}
		p.RecordHesitation(1000)
		assert.Equal(t, 1, p.HesitationCount)
		assert.InDelta(t, 1000.0, p.HesitationMean, 0.001)
		assert.InDelta(t, 0.0, p.HesitationM2, 0.001)
	})

	t.Run("Welford running stats", func(t *testing.T) {
		p := &BettingHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		assert.Equal(t, 3, p.HesitationCount)
		assert.InDelta(t, 2000.0, p.HesitationMean, 0.001)
	})
}

func TestBettingHumanProfile_HesitationStdDev(t *testing.T) {
	t.Run("fewer than 2 data points returns 0", func(t *testing.T) {
		p := &BettingHumanProfile{}
		assert.Equal(t, 0.0, p.HesitationStdDev())

		p.RecordHesitation(1000)
		assert.Equal(t, 0.0, p.HesitationStdDev())
	})

	t.Run("computed from data", func(t *testing.T) {
		p := &BettingHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		assert.InDelta(t, 1000.0, p.HesitationStdDev(), 1.0)
	})
}

func TestBettingHumanProfile_HesitationZScore(t *testing.T) {
	t.Run("no data returns 0", func(t *testing.T) {
		p := &BettingHumanProfile{}
		assert.Equal(t, 0.0, p.HesitationZScore(5000))
	})

	t.Run("stddev 0 returns 0", func(t *testing.T) {
		p := &BettingHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(1000)
		assert.Equal(t, 0.0, p.HesitationZScore(5000))
	})

	t.Run("computed z-score", func(t *testing.T) {
		p := &BettingHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		sd := p.HesitationStdDev()
		expected := 3000.0 / sd
		assert.InDelta(t, expected, p.HesitationZScore(5000), 0.001)
	})
}

func TestBettingHumanProfile_HesitationBoost(t *testing.T) {
	t.Run("fewer than hesitationMinPlays returns 0", func(t *testing.T) {
		p := &BettingHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		assert.Equal(t, 0.0, p.HesitationBoost(10000))
	})

	t.Run("z-score below threshold returns 0", func(t *testing.T) {
		p := &BettingHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		assert.Equal(t, 0.0, p.HesitationBoost(2000))
	})

	t.Run("z-score above threshold produces boost", func(t *testing.T) {
		p := &BettingHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		// mean=2000, sd=1000, z(5000)=3.0, boost=(3.0-1.0)*0.05=0.10 → at cap
		assert.InDelta(t, 0.10, p.HesitationBoost(5000), 0.001)
	})

	t.Run("moderate hesitation gives partial boost", func(t *testing.T) {
		p := &BettingHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		sd := p.HesitationStdDev()
		ms := int(2000.0 + 2.0*sd) // z=2.0
		// boost = (2.0-1.0)*0.05 = 0.05
		assert.InDelta(t, 0.05, p.HesitationBoost(ms), 0.01)
	})
}

func TestBettingHumanProfile_AdjustedCallChance(t *testing.T) {
	// No data, no adapt → base unchanged
	p := &BettingHumanProfile{}
	assert.InDelta(t, 0.3, p.AdjustedCallChance(0.3, 0, 0), 0.001)

	// High bluff rate, max adapt
	p2 := &BettingHumanProfile{GamesPlayed: 5}
	p2.AggressiveByBracket[0] = struct{ Aggressive, Total int }{8, 10} // 80% aggressive rate
	// 0.3 + (0.8-0.5)*0.2 = 0.3 + 0.06 = 0.36
	assert.InDelta(t, 0.36, p2.AdjustedCallChance(0.3, 0, 0), 0.001)

	// Low bluff rate, max adapt
	p3 := &BettingHumanProfile{GamesPlayed: 5}
	p3.AggressiveByBracket[1] = struct{ Aggressive, Total int }{1, 10} // 10% rate
	// 0.3 + (0.1-0.5)*0.2 = 0.3 - 0.08 = 0.22
	assert.InDelta(t, 0.22, p3.AdjustedCallChance(0.3, 1, 0), 0.001)

	// No data bracket → 0.5, no change
	p4 := &BettingHumanProfile{GamesPlayed: 5}
	assert.InDelta(t, 0.3, p4.AdjustedCallChance(0.3, 2, 0), 0.001)
}

func TestBettingHumanProfile_AdjustedCallChance_WithHesitation(t *testing.T) {
	p := &BettingHumanProfile{GamesPlayed: 5}
	p.AggressiveByBracket[0] = struct{ Aggressive, Total int }{8, 10} // 80%
	p.RecordHesitation(1000)
	p.RecordHesitation(2000)
	p.RecordHesitation(3000)
	// bluff term: (0.8-0.5)*0.2 = 0.06
	// hesitation at 5000ms: boost capped at 0.10, * adaptStrength 0.2 = 0.02
	// total: 0.3 + 0.06 + 0.02 = 0.38
	assert.InDelta(t, 0.38, p.AdjustedCallChance(0.3, 0, 5000), 0.001)

	// No hesitation data (humanPlayMs=0) → no hesitation boost
	p2 := &BettingHumanProfile{GamesPlayed: 5}
	p2.AggressiveByBracket[0] = struct{ Aggressive, Total int }{8, 10}
	assert.InDelta(t, 0.36, p2.AdjustedCallChance(0.3, 0, 0), 0.001)
}

func TestBettingHumanProfile_AdjustedBluffChance(t *testing.T) {
	// No data → base unchanged (FoldRate=0.5, adapt=0.0)
	p := &BettingHumanProfile{}
	assert.InDelta(t, 0.4, p.AdjustedBluffChance(0.4), 0.001)

	// High fold rate, max adapt → CPU bluffs more
	p2 := &BettingHumanProfile{GamesPlayed: 5, FoldToBetCount: 9, FoldToBetTotal: 10}
	// 0.4 * (1.0 + (0.9-0.5)*0.2) = 0.4 * (1.0+0.08) = 0.4 * 1.08 = 0.432
	assert.InDelta(t, 0.432, p2.AdjustedBluffChance(0.4), 0.001)

	// Low fold rate, max adapt → CPU bluffs less
	p3 := &BettingHumanProfile{GamesPlayed: 5, FoldToBetCount: 1, FoldToBetTotal: 10}
	// 0.4 * (1.0 + (0.1-0.5)*0.2) = 0.4 * (1.0-0.08) = 0.4 * 0.92 = 0.368
	assert.InDelta(t, 0.368, p3.AdjustedBluffChance(0.4), 0.001)

	// Neutral fold rate → no change
	p4 := &BettingHumanProfile{GamesPlayed: 5, FoldToBetCount: 5, FoldToBetTotal: 10}
	assert.InDelta(t, 0.4, p4.AdjustedBluffChance(0.4), 0.001)
}
