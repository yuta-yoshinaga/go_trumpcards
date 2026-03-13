//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPokerConfig_Validate(t *testing.T) {
	validCfg := func() domain.PokerConfig {
		return domain.PokerConfig{
			BettingLimit: domain.BettingLimitFixed,
			CpuCount:     1,
			JokerCount:   0,
		}
	}
	t.Run("valid config returns nil", func(t *testing.T) {
		assert.NoError(t, validCfg().Validate())
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
	t.Run("joker count negative returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.JokerCount = -1
		assert.Error(t, cfg.Validate())
	})
	t.Run("joker count above max returns error", func(t *testing.T) {
		cfg := validCfg()
		cfg.JokerCount = 3
		assert.Error(t, cfg.Validate())
	})
}
