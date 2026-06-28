//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultPrsiConfig(t *testing.T) {
	cfg := domain.DefaultPrsiConfig()
	assert.Equal(t, domain.PrsiCpuDifficultyNormal, cfg.CpuDifficulty)
}

func TestPrsiConfig_Validate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		assert.NoError(t, domain.DefaultPrsiConfig().Validate())
	})
	t.Run("valid easy difficulty", func(t *testing.T) {
		cfg := domain.DefaultPrsiConfig()
		cfg.CpuDifficulty = domain.PrsiCpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})
	t.Run("valid hard difficulty", func(t *testing.T) {
		cfg := domain.DefaultPrsiConfig()
		cfg.CpuDifficulty = domain.PrsiCpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})
	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultPrsiConfig()
		cfg.CpuDifficulty = domain.PrsiCpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})
	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultPrsiConfig()
		cfg.CpuDifficulty = domain.PrsiCpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})
}
