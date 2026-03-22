//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndianPokerCardBracket(t *testing.T) {
	tests := []struct {
		name     string
		cardRank int
		want     int
	}{
		{"rank 2 → weak(0)", 2, 0},
		{"rank 5 → weak(0)", 5, 0},
		{"rank 6 → medium(1)", 6, 1},
		{"rank 9 → medium(1)", 9, 1},
		{"rank 10 → strong(2)", 10, 2},
		{"rank 14 (Ace) → strong(2)", 14, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, indianPokerCardBracket(tt.cardRank))
		})
	}
}

func TestIndianPokerHumanProfile_RecordAction(t *testing.T) {
	p := &IndianPokerHumanProfile{}

	// Bet with weak card (aggressive)
	p.RecordAction(3, bettingActionBet)
	assert.Equal(t, 1, p.AggressiveByBracket[0].Aggressive)
	assert.Equal(t, 1, p.AggressiveByBracket[0].Total)

	// Raise with weak card (aggressive)
	p.RecordAction(4, bettingActionRaise)
	assert.Equal(t, 2, p.AggressiveByBracket[0].Aggressive)
	assert.Equal(t, 2, p.AggressiveByBracket[0].Total)

	// Call with weak card (not aggressive)
	p.RecordAction(2, bettingActionCall)
	assert.Equal(t, 2, p.AggressiveByBracket[0].Aggressive)
	assert.Equal(t, 3, p.AggressiveByBracket[0].Total)

	// Check with medium card (not aggressive)
	p.RecordAction(7, bettingActionCheck)
	assert.Equal(t, 0, p.AggressiveByBracket[1].Aggressive)
	assert.Equal(t, 1, p.AggressiveByBracket[1].Total)

	// Fold with strong card (not aggressive)
	p.RecordAction(12, bettingActionFold)
	assert.Equal(t, 0, p.AggressiveByBracket[2].Aggressive)
	assert.Equal(t, 1, p.AggressiveByBracket[2].Total)
}

func TestIndianPokerHumanProfile_RecordFoldToBet(t *testing.T) {
	p := &IndianPokerHumanProfile{}

	p.RecordFoldToBet(true)
	assert.Equal(t, 1, p.FoldToBetCount)
	assert.Equal(t, 1, p.FoldToBetTotal)

	p.RecordFoldToBet(false)
	assert.Equal(t, 1, p.FoldToBetCount)
	assert.Equal(t, 2, p.FoldToBetTotal)
}

func TestIndianPokerHumanProfile_RecordHesitation(t *testing.T) {
	t.Run("ms=0 is no-op", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(0)
		assert.Equal(t, 0, p.HesitationCount)
		assert.Equal(t, 0.0, p.HesitationMean)
	})

	t.Run("ms<0 is no-op", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(-5)
		assert.Equal(t, 0, p.HesitationCount)
	})

	t.Run("ms>0 records data", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		assert.Equal(t, 1, p.HesitationCount)
		assert.InDelta(t, 1000.0, p.HesitationMean, 0.001)
		assert.InDelta(t, 0.0, p.HesitationM2, 0.001)
	})

	t.Run("multiple values Welford update", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		assert.Equal(t, 3, p.HesitationCount)
		assert.InDelta(t, 2000.0, p.HesitationMean, 0.001)
	})
}

func TestIndianPokerHumanProfile_BluffRate(t *testing.T) {
	p := &IndianPokerHumanProfile{}

	// No data → 0.5
	assert.Equal(t, 0.5, p.BluffRate(0))
	assert.Equal(t, 0.5, p.BluffRate(1))
	assert.Equal(t, 0.5, p.BluffRate(2))

	// Invalid bracket → 0.5
	assert.Equal(t, 0.5, p.BluffRate(-1))
	assert.Equal(t, 0.5, p.BluffRate(3))

	// With data
	p.RecordAction(3, bettingActionBet)  // bracket 0, aggressive
	p.RecordAction(3, bettingActionCall) // bracket 0, not aggressive
	assert.Equal(t, 0.5, p.BluffRate(0)) // 1/2 = 0.5

	p.RecordAction(2, bettingActionRaise) // bracket 0, aggressive
	assert.InDelta(t, 2.0/3.0, p.BluffRate(0), 0.001)
}

func TestIndianPokerHumanProfile_FoldRate(t *testing.T) {
	p := &IndianPokerHumanProfile{}

	// No data → 0.5
	assert.Equal(t, 0.5, p.FoldRate())

	// With data
	p.RecordFoldToBet(true)
	p.RecordFoldToBet(true)
	p.RecordFoldToBet(false)
	assert.InDelta(t, 2.0/3.0, p.FoldRate(), 0.001)
}

func TestIndianPokerHumanProfile_AdaptStrength(t *testing.T) {
	tests := []struct {
		name        string
		gamesPlayed int
		want        float64
	}{
		{"0 games", 0, 0.0},
		{"3 games", 3, 0.12},
		{"5 games", 5, 0.20},
		{"10 games (capped at 5)", 10, 0.20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &IndianPokerHumanProfile{GamesPlayed: tt.gamesPlayed}
			assert.InDelta(t, tt.want, p.AdaptStrength(), 0.001)
		})
	}
}

func TestIndianPokerHumanProfile_HesitationStdDev(t *testing.T) {
	t.Run("no data returns 0", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		assert.Equal(t, 0.0, p.HesitationStdDev())
	})

	t.Run("one data point returns 0", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		assert.Equal(t, 0.0, p.HesitationStdDev())
	})

	t.Run("computed from data", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		assert.InDelta(t, 1000.0, p.HesitationStdDev(), 1.0)
	})
}

func TestIndianPokerHumanProfile_HesitationZScore(t *testing.T) {
	t.Run("stddev=0 returns 0", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		assert.Equal(t, 0.0, p.HesitationZScore(5000))
	})

	t.Run("same values stddev=0 returns 0", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(1000)
		assert.Equal(t, 0.0, p.HesitationZScore(5000))
	})

	t.Run("stddev>0 returns correct z", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		sd := p.HesitationStdDev()
		// mean=2000, z = (5000-2000)/sd = 3000/sd
		expected := 3000.0 / sd
		assert.InDelta(t, expected, p.HesitationZScore(5000), 0.001)
	})
}

func TestIndianPokerHumanProfile_HesitationBoost(t *testing.T) {
	t.Run("fewer than hesitationMinPlays(3) returns 0", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		assert.Equal(t, 0.0, p.HesitationBoost(10000))
	})

	t.Run("z<=1.0 returns 0", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		// mean=2000, sd=1000, z(2000)=0.0
		assert.Equal(t, 0.0, p.HesitationBoost(2000))
	})

	t.Run("z>1.0 produces boost", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		sd := p.HesitationStdDev()
		ms := int(2000.0 + 2.0*sd) // z=2.0
		// boost = (2.0 - 1.0) * 0.05 = 0.05
		assert.InDelta(t, 0.05, p.HesitationBoost(ms), 0.01)
	})

	t.Run("capped at maxHesitationBoost(0.10)", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		// mean=2000, sd=1000, z(5000)=3.0, boost=(3.0-1.0)*0.05=0.10 → at cap
		assert.InDelta(t, 0.10, p.HesitationBoost(5000), 0.001)
	})

	t.Run("exceeding cap still returns max", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		// z(10000)=8.0, boost=(8.0-1.0)*0.05=0.35 → capped at 0.10
		assert.InDelta(t, 0.10, p.HesitationBoost(10000), 0.001)
	})
}

func TestIndianPokerHumanProfile_AdjustedCallChance(t *testing.T) {
	t.Run("no data → base unchanged", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		assert.InDelta(t, 0.3, p.AdjustedCallChance(0.3, 0, 0), 0.001)
	})

	t.Run("high bluff rate max adapt", func(t *testing.T) {
		p := &IndianPokerHumanProfile{GamesPlayed: 5}
		p.AggressiveByBracket[0] = struct{ Aggressive, Total int }{8, 10} // 80%
		// 0.3 + (0.8-0.5)*0.2 = 0.3 + 0.06 = 0.36
		assert.InDelta(t, 0.36, p.AdjustedCallChance(0.3, 0, 0), 0.001)
	})

	t.Run("with hesitation boost", func(t *testing.T) {
		p := &IndianPokerHumanProfile{GamesPlayed: 5}
		p.AggressiveByBracket[0] = struct{ Aggressive, Total int }{8, 10}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		// bluff term: (0.8-0.5)*0.2 = 0.06
		// hesitation at 5000ms: boost capped at 0.10, * adapt 0.2 = 0.02
		// total: 0.3 + 0.06 + 0.02 = 0.38
		assert.InDelta(t, 0.38, p.AdjustedCallChance(0.3, 0, 5000), 0.001)
	})

	t.Run("invalid bracket → bluff rate 0.5, no change", func(t *testing.T) {
		p := &IndianPokerHumanProfile{GamesPlayed: 5}
		assert.InDelta(t, 0.3, p.AdjustedCallChance(0.3, 2, 0), 0.001)
	})
}

func TestIndianPokerHumanProfile_AdjustedBluffChance(t *testing.T) {
	t.Run("no data → base unchanged", func(t *testing.T) {
		p := &IndianPokerHumanProfile{}
		assert.InDelta(t, 0.4, p.AdjustedBluffChance(0.4), 0.001)
	})

	t.Run("high fold rate → CPU bluffs more", func(t *testing.T) {
		p := &IndianPokerHumanProfile{GamesPlayed: 5, FoldToBetCount: 9, FoldToBetTotal: 10}
		// 0.4 * (1.0 + (0.9-0.5)*0.2) = 0.4 * 1.08 = 0.432
		assert.InDelta(t, 0.432, p.AdjustedBluffChance(0.4), 0.001)
	})

	t.Run("low fold rate → CPU bluffs less", func(t *testing.T) {
		p := &IndianPokerHumanProfile{GamesPlayed: 5, FoldToBetCount: 1, FoldToBetTotal: 10}
		// 0.4 * (1.0 + (0.1-0.5)*0.2) = 0.4 * 0.92 = 0.368
		assert.InDelta(t, 0.368, p.AdjustedBluffChance(0.4), 0.001)
	})

	t.Run("neutral fold rate → no change", func(t *testing.T) {
		p := &IndianPokerHumanProfile{GamesPlayed: 5, FoldToBetCount: 5, FoldToBetTotal: 10}
		assert.InDelta(t, 0.4, p.AdjustedBluffChance(0.4), 0.001)
	})
}
