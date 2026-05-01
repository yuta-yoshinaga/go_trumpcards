//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultTonkConfig(t *testing.T) {
	cfg := domain.DefaultTonkConfig()
	assert.Equal(t, domain.TonkCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 50, cfg.PointLimit)
}

func TestTonkConfig_Validate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		cfg := domain.DefaultTonkConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid easy difficulty", func(t *testing.T) {
		cfg := domain.DefaultTonkConfig()
		cfg.CpuDifficulty = domain.TonkCpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid hard difficulty", func(t *testing.T) {
		cfg := domain.DefaultTonkConfig()
		cfg.CpuDifficulty = domain.TonkCpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})

	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultTonkConfig()
		cfg.CpuDifficulty = domain.TonkCpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultTonkConfig()
		cfg.CpuDifficulty = domain.TonkCpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("point limit zero", func(t *testing.T) {
		cfg := domain.DefaultTonkConfig()
		cfg.PointLimit = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit 1 is valid", func(t *testing.T) {
		cfg := domain.DefaultTonkConfig()
		cfg.PointLimit = 1
		assert.NoError(t, cfg.Validate())
	})
}
