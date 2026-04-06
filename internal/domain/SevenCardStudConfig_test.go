//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSevenCardStudConfig(t *testing.T) {
	cfg := DefaultSevenCardStudConfig()
	assert.Equal(t, 1, cfg.Ante)
	assert.Equal(t, 2, cfg.BringIn)
	assert.Equal(t, 5, cfg.SmallBet)
	assert.Equal(t, 10, cfg.BigBet)
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, BettingLimitFixed, cfg.BettingLimit)
	assert.Equal(t, SevenCardStudTableSize7, cfg.TableSize)
	assert.False(t, cfg.TournamentMode)
	assert.NoError(t, cfg.Validate())
}

func TestSevenCardStudConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*SevenCardStudConfig)
		wantErr bool
	}{
		{name: "valid default", modify: func(c *SevenCardStudConfig) {}, wantErr: false},
		{name: "invalid betting limit", modify: func(c *SevenCardStudConfig) { c.BettingLimit = 99 }, wantErr: true},
		{name: "ante zero", modify: func(c *SevenCardStudConfig) { c.Ante = 0 }, wantErr: true},
		{name: "bring-in zero", modify: func(c *SevenCardStudConfig) { c.BringIn = 0 }, wantErr: true},
		{name: "small bet zero", modify: func(c *SevenCardStudConfig) { c.SmallBet = 0 }, wantErr: true},
		{name: "big bet zero", modify: func(c *SevenCardStudConfig) { c.BigBet = 0 }, wantErr: true},
		{name: "small bet > big bet", modify: func(c *SevenCardStudConfig) { c.SmallBet = 20; c.BigBet = 10 }, wantErr: true},
		{name: "small bet == big bet", modify: func(c *SevenCardStudConfig) { c.SmallBet = 10; c.BigBet = 10 }, wantErr: false},
		{name: "ante level hands zero", modify: func(c *SevenCardStudConfig) { c.AnteLevelHands = 0 }, wantErr: true},
		{name: "table size 1", modify: func(c *SevenCardStudConfig) { c.TableSize = 1 }, wantErr: true},
		{name: "table size 8", modify: func(c *SevenCardStudConfig) { c.TableSize = 8 }, wantErr: true},
		{name: "table size 2", modify: func(c *SevenCardStudConfig) { c.TableSize = 2 }, wantErr: false},
		{name: "table size 0 (no change)", modify: func(c *SevenCardStudConfig) { c.TableSize = 0 }, wantErr: false},
		{name: "pot limit", modify: func(c *SevenCardStudConfig) { c.BettingLimit = BettingLimitPotLimit }, wantErr: false},
		{name: "no limit", modify: func(c *SevenCardStudConfig) { c.BettingLimit = BettingLimitNoLimit }, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultSevenCardStudConfig()
			tt.modify(&cfg)
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidSevenCardStudTableSize(t *testing.T) {
	tests := []struct {
		size int
		want bool
	}{
		{1, false},
		{2, true},
		{3, true},
		{4, true},
		{5, true},
		{6, true},
		{7, true},
		{8, false},
		{0, false},
		{-1, false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, IsValidSevenCardStudTableSize(tt.size), "size=%d", tt.size)
	}
}

func TestDefaultSevenCardStudCpuStyles(t *testing.T) {
	assert.Len(t, DefaultSevenCardStudCpuStyles(2), 1)
	assert.Len(t, DefaultSevenCardStudCpuStyles(3), 2)
	assert.Len(t, DefaultSevenCardStudCpuStyles(4), 3)
	assert.Len(t, DefaultSevenCardStudCpuStyles(5), 4)
	assert.Len(t, DefaultSevenCardStudCpuStyles(6), 5)
	assert.Len(t, DefaultSevenCardStudCpuStyles(7), 6)
	// Edge: size <= 1
	assert.Len(t, DefaultSevenCardStudCpuStyles(1), 1)
	// Edge: size > 7
	assert.Len(t, DefaultSevenCardStudCpuStyles(8), 6)
}

func TestNewSevenCardStudPlayersForTable(t *testing.T) {
	tests := []struct {
		size     int
		wantLen  int
		wantSize int // expected actual table size
	}{
		{7, 7, 7},
		{4, 4, 4},
		{2, 2, 2},
		{1, 7, 7},  // invalid -> default 7
		{8, 7, 7},  // invalid -> default 7
		{-1, 7, 7}, // invalid -> default 7
	}
	for _, tt := range tests {
		players := NewSevenCardStudPlayersForTable(tt.size)
		assert.Len(t, players, tt.wantLen, "size=%d", tt.size)
		assert.True(t, players[0].GetIsHuman(), "first player should be human")
		for i := 1; i < len(players); i++ {
			assert.False(t, players[i].GetIsHuman(), "player %d should be CPU", i)
		}
	}
}
