package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOldMaidPickBucket(t *testing.T) {
	tests := []struct {
		name     string
		cardIdx  int
		handSize int
		want     int
	}{
		{"idx 0 of 9 → start", 0, 9, 0},
		{"idx 2 of 9 → start", 2, 9, 0},
		{"idx 3 of 9 → middle", 3, 9, 1},
		{"idx 5 of 9 → middle", 5, 9, 1},
		{"idx 6 of 9 → end", 6, 9, 2},
		{"idx 8 of 9 → end", 8, 9, 2},
		{"idx 0 of 1 → start", 0, 1, 0},
		{"idx 0 of 2 → start", 0, 2, 0},
		{"idx 1 of 2 → middle", 1, 2, 1},
		{"handSize 0 → start", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, oldMaidPickBucket(tt.cardIdx, tt.handSize))
		})
	}
}

func TestOldMaidHumanProfile_RecordPick(t *testing.T) {
	p := &OldMaidHumanProfile{}

	p.RecordPick(0, 9)
	assert.Equal(t, 1, p.PositionBuckets[0])
	assert.Equal(t, 1, p.TotalPicks)

	p.RecordPick(4, 9)
	assert.Equal(t, 1, p.PositionBuckets[1])
	assert.Equal(t, 2, p.TotalPicks)

	p.RecordPick(8, 9)
	assert.Equal(t, 1, p.PositionBuckets[2])
	assert.Equal(t, 3, p.TotalPicks)
}

func TestOldMaidHumanProfile_RecordShuffleAndDraw(t *testing.T) {
	p := &OldMaidHumanProfile{}

	p.RecordShuffle()
	assert.Equal(t, 1, p.ShuffleCount)

	p.RecordDraw()
	p.RecordDraw()
	assert.Equal(t, 2, p.DrawCount)
}

func TestOldMaidHumanProfile_PickRate(t *testing.T) {
	p := &OldMaidHumanProfile{}

	// No data → 1/3
	assert.InDelta(t, 1.0/3.0, p.PickRate(0), 0.001)
	assert.InDelta(t, 1.0/3.0, p.PickRate(1), 0.001)
	assert.InDelta(t, 1.0/3.0, p.PickRate(2), 0.001)

	// Out of range → 1/3
	assert.InDelta(t, 1.0/3.0, p.PickRate(-1), 0.001)
	assert.InDelta(t, 1.0/3.0, p.PickRate(3), 0.001)

	// With data
	p.RecordPick(0, 9) // bucket 0
	p.RecordPick(0, 9) // bucket 0
	p.RecordPick(4, 9) // bucket 1
	assert.InDelta(t, 2.0/3.0, p.PickRate(0), 0.001)
	assert.InDelta(t, 1.0/3.0, p.PickRate(1), 0.001)
	assert.InDelta(t, 0.0, p.PickRate(2), 0.001)
}

func TestOldMaidHumanProfile_ShuffleRate(t *testing.T) {
	p := &OldMaidHumanProfile{}

	// No draws → 0.0
	assert.Equal(t, 0.0, p.ShuffleRate())

	// With data
	p.DrawCount = 10
	p.ShuffleCount = 3
	assert.InDelta(t, 0.3, p.ShuffleRate(), 0.001)
}

func TestOldMaidHumanProfile_AdaptStrength(t *testing.T) {
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
			p := &OldMaidHumanProfile{GamesPlayed: tt.gamesPlayed}
			assert.InDelta(t, tt.want, p.AdaptStrength(), 0.001)
		})
	}
}

func TestOldMaidHumanProfile_StrategicPlacement(t *testing.T) {
	t.Run("size <= 1 returns 0", func(t *testing.T) {
		p := &OldMaidHumanProfile{GamesPlayed: 5}
		assert.Equal(t, 0, p.StrategicPlacement(0))
		assert.Equal(t, 0, p.StrategicPlacement(1))
	})

	t.Run("least-picked bucket 0 → position 0", func(t *testing.T) {
		p := &OldMaidHumanProfile{
			GamesPlayed:     5,
			TotalPicks:      30,
			PositionBuckets: [3]int{2, 14, 14}, // bucket 0 least picked
		}
		assert.Equal(t, 0, p.StrategicPlacement(9))
	})

	t.Run("least-picked bucket 2 → end position", func(t *testing.T) {
		p := &OldMaidHumanProfile{
			GamesPlayed:     5,
			TotalPicks:      30,
			PositionBuckets: [3]int{14, 14, 2}, // bucket 2 least picked
		}
		assert.Equal(t, 8, p.StrategicPlacement(9))
	})

	t.Run("least-picked bucket 1 → middle position", func(t *testing.T) {
		p := &OldMaidHumanProfile{
			GamesPlayed:     5,
			TotalPicks:      30,
			PositionBuckets: [3]int{14, 2, 14}, // bucket 1 least picked
		}
		pos := p.StrategicPlacement(9)
		// third = 3, so position should be in [3, 5]
		assert.GreaterOrEqual(t, pos, 3)
		assert.Less(t, pos, 6)
	})
}
