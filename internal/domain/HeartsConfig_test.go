//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultHeartsConfig(t *testing.T) {
	cfg := domain.DefaultHeartsConfig()
	assert.Equal(t, domain.HeartsCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 100, cfg.PointLimit)
}

func TestHeartsCpuDifficultyConstants(t *testing.T) {
	assert.Equal(t, domain.HeartsCpuDifficulty(0), domain.HeartsCpuDifficultyEasy)
	assert.Equal(t, domain.HeartsCpuDifficulty(1), domain.HeartsCpuDifficultyNormal)
	assert.Equal(t, domain.HeartsCpuDifficulty(2), domain.HeartsCpuDifficultyHard)
}

func TestHeartsConfig_Validate(t *testing.T) {
	t.Run("valid config returns nil", func(t *testing.T) {
		cfg := domain.HeartsConfig{CpuDifficulty: domain.HeartsCpuDifficultyNormal, PointLimit: 100}
		assert.NoError(t, cfg.Validate())
	})
	t.Run("cpu difficulty below min returns error", func(t *testing.T) {
		cfg := domain.HeartsConfig{CpuDifficulty: domain.HeartsCpuDifficulty(-1), PointLimit: 100}
		assert.Error(t, cfg.Validate())
	})
	t.Run("cpu difficulty above max returns error", func(t *testing.T) {
		cfg := domain.HeartsConfig{CpuDifficulty: domain.HeartsCpuDifficulty(99), PointLimit: 100}
		assert.Error(t, cfg.Validate())
	})
	t.Run("point limit zero returns error", func(t *testing.T) {
		cfg := domain.HeartsConfig{CpuDifficulty: domain.HeartsCpuDifficultyNormal, PointLimit: 0}
		assert.Error(t, cfg.Validate())
	})
	t.Run("point limit negative returns error", func(t *testing.T) {
		cfg := domain.HeartsConfig{CpuDifficulty: domain.HeartsCpuDifficultyNormal, PointLimit: -1}
		assert.Error(t, cfg.Validate())
	})
}
