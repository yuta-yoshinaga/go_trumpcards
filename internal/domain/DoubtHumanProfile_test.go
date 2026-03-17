package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoubtHandSizeBracket(t *testing.T) {
	tests := []struct {
		name     string
		handSize int
		want     int
	}{
		{"1 card → small", 1, 0},
		{"4 cards → small", 4, 0},
		{"5 cards → medium", 5, 1},
		{"9 cards → medium", 9, 1},
		{"10 cards → large", 10, 2},
		{"20 cards → large", 20, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, doubtHandSizeBracket(tt.handSize))
		})
	}
}

func TestDoubtHumanProfile_RecordPlay(t *testing.T) {
	p := &DoubtHumanProfile{}

	// Bluff in small bracket
	p.RecordPlay(3, true)
	assert.Equal(t, 1, p.BluffsByBracket[0].Bluffs)
	assert.Equal(t, 1, p.BluffsByBracket[0].Total)

	// Honest in small bracket
	p.RecordPlay(2, false)
	assert.Equal(t, 1, p.BluffsByBracket[0].Bluffs)
	assert.Equal(t, 2, p.BluffsByBracket[0].Total)

	// Bluff in medium bracket
	p.RecordPlay(7, true)
	assert.Equal(t, 1, p.BluffsByBracket[1].Bluffs)
	assert.Equal(t, 1, p.BluffsByBracket[1].Total)

	// Bluff in large bracket
	p.RecordPlay(12, true)
	assert.Equal(t, 1, p.BluffsByBracket[2].Bluffs)
	assert.Equal(t, 1, p.BluffsByBracket[2].Total)
}

func TestDoubtHumanProfile_RecordDoubt(t *testing.T) {
	p := &DoubtHumanProfile{}

	p.RecordDoubt(true)
	assert.Equal(t, 1, p.DoubtCorrect)
	assert.Equal(t, 1, p.DoubtTotal)

	p.RecordDoubt(false)
	assert.Equal(t, 1, p.DoubtCorrect)
	assert.Equal(t, 2, p.DoubtTotal)
}

func TestDoubtHumanProfile_BluffRate(t *testing.T) {
	p := &DoubtHumanProfile{}

	// No data → 0.5
	assert.Equal(t, 0.5, p.BluffRate(0))
	assert.Equal(t, 0.5, p.BluffRate(1))
	assert.Equal(t, 0.5, p.BluffRate(2))

	// Out of range → 0.5
	assert.Equal(t, 0.5, p.BluffRate(-1))
	assert.Equal(t, 0.5, p.BluffRate(3))

	// With data
	p.RecordPlay(3, true)
	p.RecordPlay(3, false)
	assert.Equal(t, 0.5, p.BluffRate(0))

	p.RecordPlay(2, true)
	assert.InDelta(t, 2.0/3.0, p.BluffRate(0), 0.001)
}

func TestDoubtHumanProfile_DoubtAccuracy(t *testing.T) {
	p := &DoubtHumanProfile{}

	// No data → 0.5
	assert.Equal(t, 0.5, p.DoubtAccuracy())

	// With data
	p.RecordDoubt(true)
	p.RecordDoubt(true)
	p.RecordDoubt(false)
	assert.InDelta(t, 2.0/3.0, p.DoubtAccuracy(), 0.001)
}

func TestDoubtHumanProfile_AdaptStrength(t *testing.T) {
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
			p := &DoubtHumanProfile{GamesPlayed: tt.gamesPlayed}
			assert.InDelta(t, tt.want, p.AdaptStrength(), 0.001)
		})
	}
}

func TestDoubtHumanProfile_AdjustedDoubtChance(t *testing.T) {
	// No data, no adapt → base unchanged
	p := &DoubtHumanProfile{}
	assert.InDelta(t, 0.3, p.AdjustedDoubtChance(0.3, 0, 0), 0.001)

	// High bluff rate, max adapt strength
	p2 := &DoubtHumanProfile{GamesPlayed: 5}
	p2.BluffsByBracket[0] = struct{ Bluffs, Total int }{8, 10} // 80% bluff rate
	// 0.3 + (0.8 - 0.5) * 0.2 = 0.3 + 0.06 = 0.36
	assert.InDelta(t, 0.36, p2.AdjustedDoubtChance(0.3, 0, 0), 0.001)

	// Low bluff rate, max adapt
	p3 := &DoubtHumanProfile{GamesPlayed: 5}
	p3.BluffsByBracket[1] = struct{ Bluffs, Total int }{1, 10} // 10% bluff rate
	// 0.3 + (0.1 - 0.5) * 0.2 = 0.3 - 0.08 = 0.22
	assert.InDelta(t, 0.22, p3.AdjustedDoubtChance(0.3, 1, 0), 0.001)

	// No data bracket → 0.5, no change
	p4 := &DoubtHumanProfile{GamesPlayed: 5}
	assert.InDelta(t, 0.3, p4.AdjustedDoubtChance(0.3, 2, 0), 0.001)
}

func TestDoubtHumanProfile_RecordHesitation(t *testing.T) {
	t.Run("ms<=0 is no-op", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		p.RecordHesitation(0)
		assert.Equal(t, 0, p.HesitationCount)
		assert.Equal(t, 0.0, p.HesitationMean)

		p.RecordHesitation(-1)
		assert.Equal(t, 0, p.HesitationCount)
	})

	t.Run("single data point", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		p.RecordHesitation(1000)
		assert.Equal(t, 1, p.HesitationCount)
		assert.InDelta(t, 1000.0, p.HesitationMean, 0.001)
		assert.InDelta(t, 0.0, p.HesitationM2, 0.001)
	})

	t.Run("Welford running stats", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		assert.Equal(t, 3, p.HesitationCount)
		assert.InDelta(t, 2000.0, p.HesitationMean, 0.001)
	})
}

func TestDoubtHumanProfile_HesitationStdDev(t *testing.T) {
	t.Run("fewer than 2 data points returns 0", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		assert.Equal(t, 0.0, p.HesitationStdDev())

		p.RecordHesitation(1000)
		assert.Equal(t, 0.0, p.HesitationStdDev())
	})

	t.Run("computed from data", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		// 1000, 2000, 3000 → mean=2000, sample stddev = sqrt(M2/(N-1)) = sqrt(2000000/2) = 1000
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		assert.InDelta(t, 1000.0, p.HesitationStdDev(), 1.0)
	})
}

func TestDoubtHumanProfile_HesitationZScore(t *testing.T) {
	t.Run("no data returns 0", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		assert.Equal(t, 0.0, p.HesitationZScore(5000))
	})

	t.Run("stddev 0 returns 0", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(1000)
		assert.Equal(t, 0.0, p.HesitationZScore(5000))
	})

	t.Run("computed z-score", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		sd := p.HesitationStdDev()
		// z = (5000 - 2000) / sd
		expected := 3000.0 / sd
		assert.InDelta(t, expected, p.HesitationZScore(5000), 0.001)
	})
}

func TestDoubtHumanProfile_HesitationBoost(t *testing.T) {
	t.Run("fewer than hesitationMinPlays returns 0", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		assert.Equal(t, 0.0, p.HesitationBoost(10000))
	})

	t.Run("z-score below threshold returns 0", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		// mean=2000, querying with mean value → z=0 < 1.0
		assert.Equal(t, 0.0, p.HesitationBoost(2000))
	})

	t.Run("z-score above threshold produces boost", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		// mean=2000, sd=1000, z(5000)=(3000/1000)=3.0
		// boost = (3.0 - 1.0) * 0.05 = 0.10 → exactly at cap
		assert.InDelta(t, 0.10, p.HesitationBoost(5000), 0.001)
	})

	t.Run("moderate hesitation gives partial boost", func(t *testing.T) {
		p := &DoubtHumanProfile{}
		p.RecordHesitation(1000)
		p.RecordHesitation(2000)
		p.RecordHesitation(3000)
		sd := p.HesitationStdDev()
		// target z = 2.0 → ms = mean + 2*sd
		ms := int(2000.0 + 2.0*sd)
		// boost = (2.0 - 1.0) * 0.05 = 0.05
		assert.InDelta(t, 0.05, p.HesitationBoost(ms), 0.01)
	})
}

func TestDoubtHumanProfile_AdjustedDoubtChance_WithHesitation(t *testing.T) {
	// Max adapt, high bluff, with hesitation boost
	p := &DoubtHumanProfile{GamesPlayed: 5}
	p.BluffsByBracket[0] = struct{ Bluffs, Total int }{8, 10} // 80% bluff rate
	p.RecordHesitation(1000)
	p.RecordHesitation(2000)
	p.RecordHesitation(3000)
	// bluff term: (0.8-0.5)*0.2 = 0.06
	// hesitation at 5000ms: boost capped at 0.10, * adaptStrength 0.2 = 0.02
	// total: 0.3 + 0.06 + 0.02 = 0.38
	assert.InDelta(t, 0.38, p.AdjustedDoubtChance(0.3, 0, 5000), 0.001)

	// No hesitation data (humanPlayMs=0) → no hesitation boost
	p2 := &DoubtHumanProfile{GamesPlayed: 5}
	p2.BluffsByBracket[0] = struct{ Bluffs, Total int }{8, 10}
	assert.InDelta(t, 0.36, p2.AdjustedDoubtChance(0.3, 0, 0), 0.001)
}

func TestDoubtHumanProfile_AdjustedBluffChance(t *testing.T) {
	// No data → base unchanged (0.5 * 0.0 = 0)
	p := &DoubtHumanProfile{}
	assert.InDelta(t, 0.4, p.AdjustedBluffChance(0.4), 0.001)

	// High doubt accuracy, max adapt
	p2 := &DoubtHumanProfile{GamesPlayed: 5, DoubtCorrect: 9, DoubtTotal: 10}
	// 0.4 * (1.0 - 0.9 * 0.2) = 0.4 * (1.0 - 0.18) = 0.4 * 0.82 = 0.328
	assert.InDelta(t, 0.328, p2.AdjustedBluffChance(0.4), 0.001)

	// Low doubt accuracy, max adapt
	p3 := &DoubtHumanProfile{GamesPlayed: 5, DoubtCorrect: 1, DoubtTotal: 10}
	// 0.4 * (1.0 - 0.1 * 0.2) = 0.4 * (1.0 - 0.02) = 0.4 * 0.98 = 0.392
	assert.InDelta(t, 0.392, p3.AdjustedBluffChance(0.4), 0.001)
}
