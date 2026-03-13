package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultOldMaidConfig(t *testing.T) {
	cfg := domain.DefaultOldMaidConfig()
	assert.Equal(t, domain.OldMaidModeNormal, cfg.Mode)
	assert.False(t, cfg.CpuPlacementStrategy)
	assert.False(t, cfg.CpuMemoryAI)
}

func TestOldMaidMode_Values(t *testing.T) {
	assert.Equal(t, domain.OldMaidMode(0), domain.OldMaidModeNormal)
	assert.Equal(t, domain.OldMaidMode(1), domain.OldMaidModeJijiNuki)
}

func TestOldMaidConfig_Validate(t *testing.T) {
	t.Run("normal mode returns nil", func(t *testing.T) {
		cfg := domain.OldMaidConfig{Mode: domain.OldMaidModeNormal}
		assert.NoError(t, cfg.Validate())
	})
	t.Run("jiji nuki mode returns nil", func(t *testing.T) {
		cfg := domain.OldMaidConfig{Mode: domain.OldMaidModeJijiNuki}
		assert.NoError(t, cfg.Validate())
	})
	t.Run("invalid mode returns error", func(t *testing.T) {
		cfg := domain.OldMaidConfig{Mode: domain.OldMaidMode(99)}
		assert.Error(t, cfg.Validate())
	})
}
