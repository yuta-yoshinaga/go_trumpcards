//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBadugiConfig_Validate(t *testing.T) {
	validCfg := func() domain.BadugiConfig {
		return domain.BadugiConfig{
			BettingLimit: domain.BettingLimitFixed,
			CpuCount:     1,
		}
	}
	t.Run("valid fixed-limit config returns nil", func(t *testing.T) {
		assert.NoError(t, validCfg().Validate())
	})
	t.Run("pot-limit is valid", func(t *testing.T) {
		cfg := validCfg()
		cfg.BettingLimit = domain.BettingLimitPotLimit
		assert.NoError(t, cfg.Validate())
	})
	t.Run("no-limit is valid", func(t *testing.T) {
		cfg := validCfg()
		cfg.BettingLimit = domain.BettingLimitNoLimit
		assert.NoError(t, cfg.Validate())
	})
	t.Run("betting limit below min returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.BettingLimit = domain.BettingLimitType(-1)
		assert.Error(t, cfg.Validate())
	})
	t.Run("betting limit above max returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.BettingLimit = domain.BettingLimitType(99)
		assert.Error(t, cfg.Validate())
	})
	t.Run("cpu count below min returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.CpuCount = 0
		assert.Error(t, cfg.Validate())
	})
	t.Run("cpu count above max returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.CpuCount = 4
		assert.Error(t, cfg.Validate())
	})
}

func TestDefaultBadugiConfig(t *testing.T) {
	cfg := domain.DefaultBadugiConfig()
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, domain.BettingLimitFixed, cfg.BettingLimit)
	assert.Equal(t, 3, cfg.CpuCount)
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, 10, cfg.Ante)
	assert.Equal(t, 10, cfg.MinBet)
}

func TestBadugiPlayStyleNames(t *testing.T) {
	// Ordering must match the constant iota order so name lookups are safe.
	assert.Equal(t, "Conservative", domain.BadugiPlayStyleNames[domain.BadugiStyleConservative])
	assert.Equal(t, "Balanced", domain.BadugiPlayStyleNames[domain.BadugiStyleBalanced])
	assert.Equal(t, "Aggressive", domain.BadugiPlayStyleNames[domain.BadugiStyleAggressive])
	assert.Equal(t, "Bluffer", domain.BadugiPlayStyleNames[domain.BadugiStyleBluffer])
}
