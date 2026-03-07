package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestSevensConfig_Method(t *testing.T) {
	t.Run("success DefaultSevensConfig returns all disabled", func(t *testing.T) {
		cfg := domain.DefaultSevensConfig()
		assert.False(t, cfg.TunnelEnabled)
		assert.Equal(t, 0, cfg.JokerCount)
		assert.False(t, cfg.CpuStrategy)
		assert.False(t, cfg.EndStopEnabled)
	})

	t.Run("success SevensConfig can be customized", func(t *testing.T) {
		cfg := domain.SevensConfig{
			TunnelEnabled: true,
			JokerCount:    2,
			CpuStrategy:   true,
		}
		assert.True(t, cfg.TunnelEnabled)
		assert.Equal(t, 2, cfg.JokerCount)
		assert.True(t, cfg.CpuStrategy)
	})

	t.Run("success DefaultSevensConfig returns MaxPasses 5", func(t *testing.T) {
		cfg := domain.DefaultSevensConfig()
		assert.Equal(t, domain.SevensMaxPasses, cfg.MaxPasses)
		assert.Equal(t, 5, cfg.MaxPasses)
	})

	t.Run("success SevensConfig MaxPasses can be set to 0 (unlimited)", func(t *testing.T) {
		cfg := domain.SevensConfig{MaxPasses: 0}
		assert.Equal(t, 0, cfg.MaxPasses)
	})

	t.Run("success SevensConfig MaxPasses can be set to custom value", func(t *testing.T) {
		cfg := domain.SevensConfig{MaxPasses: 3}
		assert.Equal(t, 3, cfg.MaxPasses)
	})

	t.Run("success SevensConfig EndStopEnabled can be set", func(t *testing.T) {
		cfg := domain.SevensConfig{EndStopEnabled: true}
		assert.True(t, cfg.EndStopEnabled)
	})
}
