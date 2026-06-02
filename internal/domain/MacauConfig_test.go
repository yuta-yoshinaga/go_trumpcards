//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultMacauConfig(t *testing.T) {
	cfg := domain.DefaultMacauConfig()
	assert.Equal(t, domain.MacauCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 200, cfg.PointLimit)
}

func TestMacauConfig_Validate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		cfg := domain.DefaultMacauConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid easy difficulty", func(t *testing.T) {
		cfg := domain.DefaultMacauConfig()
		cfg.CpuDifficulty = domain.MacauCpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid hard difficulty", func(t *testing.T) {
		cfg := domain.DefaultMacauConfig()
		cfg.CpuDifficulty = domain.MacauCpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})

	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultMacauConfig()
		cfg.CpuDifficulty = domain.MacauCpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultMacauConfig()
		cfg.CpuDifficulty = domain.MacauCpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("point limit zero", func(t *testing.T) {
		cfg := domain.DefaultMacauConfig()
		cfg.PointLimit = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit negative", func(t *testing.T) {
		cfg := domain.DefaultMacauConfig()
		cfg.PointLimit = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit 1 is valid", func(t *testing.T) {
		cfg := domain.DefaultMacauConfig()
		cfg.PointLimit = 1
		assert.NoError(t, cfg.Validate())
	})
}
