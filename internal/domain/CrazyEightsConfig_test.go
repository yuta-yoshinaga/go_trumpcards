//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultCrazyEightsConfig(t *testing.T) {
	cfg := domain.DefaultCrazyEightsConfig()
	assert.Equal(t, domain.CrazyEightsCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 200, cfg.PointLimit)
}

func TestCrazyEightsConfig_Validate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		cfg := domain.DefaultCrazyEightsConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid easy difficulty", func(t *testing.T) {
		cfg := domain.DefaultCrazyEightsConfig()
		cfg.CpuDifficulty = domain.CrazyEightsCpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid hard difficulty", func(t *testing.T) {
		cfg := domain.DefaultCrazyEightsConfig()
		cfg.CpuDifficulty = domain.CrazyEightsCpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})

	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultCrazyEightsConfig()
		cfg.CpuDifficulty = domain.CrazyEightsCpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultCrazyEightsConfig()
		cfg.CpuDifficulty = domain.CrazyEightsCpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("point limit zero", func(t *testing.T) {
		cfg := domain.DefaultCrazyEightsConfig()
		cfg.PointLimit = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit negative", func(t *testing.T) {
		cfg := domain.DefaultCrazyEightsConfig()
		cfg.PointLimit = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit 1 is valid", func(t *testing.T) {
		cfg := domain.DefaultCrazyEightsConfig()
		cfg.PointLimit = 1
		assert.NoError(t, cfg.Validate())
	})
}
