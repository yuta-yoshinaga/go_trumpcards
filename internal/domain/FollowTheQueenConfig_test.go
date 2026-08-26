//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFollowTheQueenConfig(t *testing.T) {
	cfg := DefaultFollowTheQueenConfig()
	assert.Equal(t, 1, cfg.Ante)
	assert.Equal(t, 2, cfg.BringIn)
	assert.Equal(t, 5, cfg.SmallBet)
	assert.Equal(t, 10, cfg.BigBet)
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, BettingLimitFixed, cfg.BettingLimit)
	assert.Equal(t, FollowTheQueenTableSize4, cfg.TableSize)
	assert.False(t, cfg.TournamentMode)
	assert.NoError(t, cfg.Validate())
}

func TestFollowTheQueenConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*FollowTheQueenConfig)
		wantErr bool
	}{
		{name: "valid default", modify: func(c *FollowTheQueenConfig) {}, wantErr: false},
		{name: "invalid betting limit", modify: func(c *FollowTheQueenConfig) { c.BettingLimit = 99 }, wantErr: true},
		{name: "ante zero", modify: func(c *FollowTheQueenConfig) { c.Ante = 0 }, wantErr: true},
		{name: "bring-in zero", modify: func(c *FollowTheQueenConfig) { c.BringIn = 0 }, wantErr: true},
		{name: "small bet zero", modify: func(c *FollowTheQueenConfig) { c.SmallBet = 0 }, wantErr: true},
		{name: "big bet zero", modify: func(c *FollowTheQueenConfig) { c.BigBet = 0 }, wantErr: true},
		{name: "small bet > big bet", modify: func(c *FollowTheQueenConfig) { c.SmallBet = 20; c.BigBet = 10 }, wantErr: true},
		{name: "small bet == big bet", modify: func(c *FollowTheQueenConfig) { c.SmallBet = 10; c.BigBet = 10 }, wantErr: false},
		{name: "ante level hands zero", modify: func(c *FollowTheQueenConfig) { c.AnteLevelHands = 0 }, wantErr: true},
		{name: "table size 1", modify: func(c *FollowTheQueenConfig) { c.TableSize = 1 }, wantErr: true},
		{name: "table size 8", modify: func(c *FollowTheQueenConfig) { c.TableSize = 8 }, wantErr: true},
		{name: "table size 2", modify: func(c *FollowTheQueenConfig) { c.TableSize = 2 }, wantErr: false},
		{name: "table size 0 (no change)", modify: func(c *FollowTheQueenConfig) { c.TableSize = 0 }, wantErr: false},
		{name: "pot limit", modify: func(c *FollowTheQueenConfig) { c.BettingLimit = BettingLimitPotLimit }, wantErr: false},
		{name: "no limit", modify: func(c *FollowTheQueenConfig) { c.BettingLimit = BettingLimitNoLimit }, wantErr: false},
		{name: "init chips zero", modify: func(c *FollowTheQueenConfig) { c.InitChips = 0 }, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultFollowTheQueenConfig()
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

func TestIsValidFollowTheQueenTableSize(t *testing.T) {
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
		assert.Equal(t, tt.want, IsValidFollowTheQueenTableSize(tt.size), "size=%d", tt.size)
	}
}

func TestDefaultFollowTheQueenCpuStyles(t *testing.T) {
	assert.Len(t, DefaultFollowTheQueenCpuStyles(2), 1)
	assert.Len(t, DefaultFollowTheQueenCpuStyles(3), 2)
	assert.Len(t, DefaultFollowTheQueenCpuStyles(4), 3)
	assert.Len(t, DefaultFollowTheQueenCpuStyles(5), 4)
	assert.Len(t, DefaultFollowTheQueenCpuStyles(6), 5)
	assert.Len(t, DefaultFollowTheQueenCpuStyles(7), 6)
	// Edge: size <= 1
	assert.Len(t, DefaultFollowTheQueenCpuStyles(1), 1)
	// Edge: size > 7
	assert.Len(t, DefaultFollowTheQueenCpuStyles(8), 6)
}

func TestNewFollowTheQueenPlayersForTable(t *testing.T) {
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
		players := NewFollowTheQueenPlayersForTable(tt.size)
		assert.Len(t, players, tt.wantLen, "size=%d", tt.size)
		assert.True(t, players[0].GetIsHuman(), "first player should be human")
		for i := 1; i < len(players); i++ {
			assert.False(t, players[i].GetIsHuman(), "player %d should be CPU", i)
		}
	}
}
