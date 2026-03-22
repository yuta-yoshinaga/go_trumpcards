//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultIndianPokerConfig(t *testing.T) {
	cfg := DefaultIndianPokerConfig()
	assert.Equal(t, 10, cfg.Ante)
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, BettingLimitNoLimit, cfg.BettingLimit)
	assert.True(t, cfg.CpuMetaAI)
}

func TestIndianPokerConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := DefaultIndianPokerConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid fixed limit", func(t *testing.T) {
		cfg := IndianPokerConfig{Ante: 1, InitChips: 1, BettingLimit: BettingLimitFixed}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid pot limit", func(t *testing.T) {
		cfg := IndianPokerConfig{Ante: 5, InitChips: 500, BettingLimit: BettingLimitPotLimit}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("ante < 1", func(t *testing.T) {
		cfg := IndianPokerConfig{Ante: 0, InitChips: 1000, BettingLimit: BettingLimitNoLimit}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ante must be >= 1")
	})

	t.Run("initChips < 1", func(t *testing.T) {
		cfg := IndianPokerConfig{Ante: 10, InitChips: 0, BettingLimit: BettingLimitNoLimit}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "init chips must be >= 1")
	})

	t.Run("betting limit below range", func(t *testing.T) {
		cfg := IndianPokerConfig{Ante: 10, InitChips: 1000, BettingLimit: BettingLimitType(-1)}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "betting limit must be")
	})

	t.Run("betting limit above range", func(t *testing.T) {
		cfg := IndianPokerConfig{Ante: 10, InitChips: 1000, BettingLimit: BettingLimitType(3)}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "betting limit must be")
	})
}

func TestNewIndianPokerPlayers(t *testing.T) {
	players := NewIndianPokerPlayers()
	assert.Len(t, players, 4)

	// Player 0: human
	assert.True(t, players[0].GetIsHuman())
	assert.Equal(t, HoldemStyleTAG, players[0].GetPlayStyle())

	// Players 1-3: CPU with TAG, LAP, TAP styles
	assert.False(t, players[1].GetIsHuman())
	assert.Equal(t, HoldemStyleTAG, players[1].GetPlayStyle())

	assert.False(t, players[2].GetIsHuman())
	assert.Equal(t, HoldemStyleLAP, players[2].GetPlayStyle())

	assert.False(t, players[3].GetIsHuman())
	assert.Equal(t, HoldemStyleTAP, players[3].GetPlayStyle())
}
