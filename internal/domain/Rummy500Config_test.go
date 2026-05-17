//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultRummy500Config(t *testing.T) {
	cfg := domain.DefaultRummy500Config()
	assert.Equal(t, domain.Rummy500CpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 500, cfg.PointLimit)
}

func TestRummy500Config_Validate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		cfg := domain.DefaultRummy500Config()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid easy difficulty", func(t *testing.T) {
		cfg := domain.DefaultRummy500Config()
		cfg.CpuDifficulty = domain.Rummy500CpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid hard difficulty", func(t *testing.T) {
		cfg := domain.DefaultRummy500Config()
		cfg.CpuDifficulty = domain.Rummy500CpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})

	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultRummy500Config()
		cfg.CpuDifficulty = domain.Rummy500CpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultRummy500Config()
		cfg.CpuDifficulty = domain.Rummy500CpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("point limit zero", func(t *testing.T) {
		cfg := domain.DefaultRummy500Config()
		cfg.PointLimit = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit negative", func(t *testing.T) {
		cfg := domain.DefaultRummy500Config()
		cfg.PointLimit = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit 1 is valid", func(t *testing.T) {
		cfg := domain.DefaultRummy500Config()
		cfg.PointLimit = 1
		assert.NoError(t, cfg.Validate())
	})
}
