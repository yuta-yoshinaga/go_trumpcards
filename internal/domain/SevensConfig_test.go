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
}
