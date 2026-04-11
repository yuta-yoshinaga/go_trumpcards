//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultTwoTenJackConfig(t *testing.T) {
	cfg := domain.DefaultTwoTenJackConfig()
	assert.Equal(t, domain.TwoTenJackCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 50, cfg.PointLimit)
}

func TestTwoTenJackConfig_Validate(t *testing.T) {
	t.Run("valid default", func(t *testing.T) {
		cfg := domain.DefaultTwoTenJackConfig()
		assert.NoError(t, cfg.Validate())
	})
	t.Run("easy difficulty ok", func(t *testing.T) {
		cfg := domain.DefaultTwoTenJackConfig()
		cfg.CpuDifficulty = domain.TwoTenJackCpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})
	t.Run("hard difficulty ok", func(t *testing.T) {
		cfg := domain.DefaultTwoTenJackConfig()
		cfg.CpuDifficulty = domain.TwoTenJackCpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})
	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultTwoTenJackConfig()
		cfg.CpuDifficulty = domain.TwoTenJackCpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})
	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultTwoTenJackConfig()
		cfg.CpuDifficulty = domain.TwoTenJackCpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})
	t.Run("point limit zero invalid", func(t *testing.T) {
		cfg := domain.DefaultTwoTenJackConfig()
		cfg.PointLimit = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})
	t.Run("point limit negative invalid", func(t *testing.T) {
		cfg := domain.DefaultTwoTenJackConfig()
		cfg.PointLimit = -5
		err := cfg.Validate()
		assert.Error(t, err)
	})
	t.Run("point limit 1 valid", func(t *testing.T) {
		cfg := domain.DefaultTwoTenJackConfig()
		cfg.PointLimit = 1
		assert.NoError(t, cfg.Validate())
	})
}
