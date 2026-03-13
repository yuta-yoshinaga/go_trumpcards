package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHoldemConfig(t *testing.T) {
	cfg := DefaultHoldemConfig()
	assert.Equal(t, 5, cfg.SmallBlind)
	assert.Equal(t, 10, cfg.BigBlind)
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, HoldemTableSize4, cfg.TableSize)
}

func TestHoldemTableSizeConstants(t *testing.T) {
	assert.Equal(t, 4, HoldemTableSize4)
	assert.Equal(t, 6, HoldemTableSize6)
	assert.Equal(t, 9, HoldemTableSize9)
}

func TestIsValidHoldemTableSize(t *testing.T) {
	assert.True(t, IsValidHoldemTableSize(4))
	assert.True(t, IsValidHoldemTableSize(6))
	assert.True(t, IsValidHoldemTableSize(9))
	assert.False(t, IsValidHoldemTableSize(2))
	assert.False(t, IsValidHoldemTableSize(5))
	assert.False(t, IsValidHoldemTableSize(10))
}

func TestDefaultCpuStyles(t *testing.T) {
	t.Run("4-max", func(t *testing.T) {
		styles := DefaultCpuStyles(HoldemTableSize4)
		assert.Equal(t, 3, len(styles))
		assert.Equal(t, HoldemStyleLAP, styles[0])
		assert.Equal(t, HoldemStyleTAP, styles[1])
		assert.Equal(t, HoldemStyleGTO, styles[2])
	})
	t.Run("6-max", func(t *testing.T) {
		styles := DefaultCpuStyles(HoldemTableSize6)
		assert.Equal(t, 5, len(styles))
		assert.Equal(t, HoldemStyleTAG, styles[0])
		assert.Equal(t, HoldemStyleLAP, styles[1])
		assert.Equal(t, HoldemStyleTAP, styles[2])
		assert.Equal(t, HoldemStyleLAG, styles[3])
		assert.Equal(t, HoldemStyleGTO, styles[4])
	})
	t.Run("9-max", func(t *testing.T) {
		styles := DefaultCpuStyles(HoldemTableSize9)
		assert.Equal(t, 8, len(styles))
	})
	t.Run("unknown defaults to 4-max", func(t *testing.T) {
		styles := DefaultCpuStyles(3)
		assert.Equal(t, 3, len(styles))
	})
}

func TestNewPlayersForTable(t *testing.T) {
	t.Run("4-max", func(t *testing.T) {
		players := NewPlayersForTable(HoldemTableSize4)
		assert.Equal(t, 4, len(players))
		assert.True(t, players[0].GetIsHuman())
		for i := 1; i < 4; i++ {
			assert.False(t, players[i].GetIsHuman())
		}
	})
	t.Run("6-max", func(t *testing.T) {
		players := NewPlayersForTable(HoldemTableSize6)
		assert.Equal(t, 6, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
	t.Run("9-max", func(t *testing.T) {
		players := NewPlayersForTable(HoldemTableSize9)
		assert.Equal(t, 9, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
	t.Run("invalid falls back to 4-max", func(t *testing.T) {
		players := NewPlayersForTable(5)
		assert.Equal(t, 4, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
}

func TestHoldemPlayStyle_Constants(t *testing.T) {
	assert.Equal(t, HoldemPlayStyle(0), HoldemStyleTAG)
	assert.Equal(t, HoldemPlayStyle(1), HoldemStyleLAP)
	assert.Equal(t, HoldemPlayStyle(2), HoldemStyleTAP)
	assert.Equal(t, HoldemPlayStyle(3), HoldemStyleLAG)
	assert.Equal(t, HoldemPlayStyle(4), HoldemStyleGTO)
}

func TestHoldemPlayStyleNames(t *testing.T) {
	assert.Equal(t, "TAG", HoldemPlayStyleNames[HoldemStyleTAG])
	assert.Equal(t, "LAP", HoldemPlayStyleNames[HoldemStyleLAP])
	assert.Equal(t, "TAP", HoldemPlayStyleNames[HoldemStyleTAP])
	assert.Equal(t, "LAG", HoldemPlayStyleNames[HoldemStyleLAG])
	assert.Equal(t, "GTO", HoldemPlayStyleNames[HoldemStyleGTO])
}

func TestHoldemConfig_Validate(t *testing.T) {
	validCfg := func() HoldemConfig {
		return HoldemConfig{
			BettingLimit:    BettingLimitFixed,
			SmallBlind:      5,
			BigBlind:        10,
			BlindLevelHands: 10,
		}
	}
	t.Run("valid config returns nil", func(t *testing.T) {
		assert.NoError(t, validCfg().Validate())
	})
	t.Run("betting limit below min returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.BettingLimit = BettingLimitType(-1)
		assert.Error(t, cfg.Validate())
	})
	t.Run("betting limit above max returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.BettingLimit = BettingLimitType(99)
		assert.Error(t, cfg.Validate())
	})
	t.Run("small blind zero returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.SmallBlind = 0
		assert.Error(t, cfg.Validate())
	})
	t.Run("big blind one returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.BigBlind = 1
		assert.Error(t, cfg.Validate())
	})
	t.Run("small blind equals big blind returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.SmallBlind = 10
		cfg.BigBlind = 10
		assert.Error(t, cfg.Validate())
	})
	t.Run("blind level hands zero returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.BlindLevelHands = 0
		assert.Error(t, cfg.Validate())
	})
	t.Run("invalid table size returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.TableSize = 5
		assert.Error(t, cfg.Validate())
	})
	t.Run("zero table size is valid (default)", func(t *testing.T) {
		cfg := validCfg()
		cfg.TableSize = 0
		assert.NoError(t, cfg.Validate())
	})
}
