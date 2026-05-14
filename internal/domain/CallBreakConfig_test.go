//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultCallBreakConfig(t *testing.T) {
	cfg := domain.DefaultCallBreakConfig()
	assert.Equal(t, domain.CallBreakCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, domain.CallBreakDefaultMaxRounds, cfg.MaxRounds)
}

func TestCallBreakConfig_Validate(t *testing.T) {
	t.Run("valid default", func(t *testing.T) {
		assert.NoError(t, domain.DefaultCallBreakConfig().Validate())
	})

	t.Run("easy difficulty valid", func(t *testing.T) {
		cfg := domain.DefaultCallBreakConfig()
		cfg.CpuDifficulty = domain.CallBreakCpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})

	t.Run("hard difficulty valid", func(t *testing.T) {
		cfg := domain.DefaultCallBreakConfig()
		cfg.CpuDifficulty = domain.CallBreakCpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})

	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultCallBreakConfig()
		cfg.CpuDifficulty = domain.CallBreakCpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultCallBreakConfig()
		cfg.CpuDifficulty = domain.CallBreakCpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("max rounds zero invalid", func(t *testing.T) {
		cfg := domain.DefaultCallBreakConfig()
		cfg.MaxRounds = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max rounds")
	})

	t.Run("max rounds negative", func(t *testing.T) {
		cfg := domain.DefaultCallBreakConfig()
		cfg.MaxRounds = -3
		err := cfg.Validate()
		assert.Error(t, err)
	})

	t.Run("max rounds 1 valid", func(t *testing.T) {
		cfg := domain.DefaultCallBreakConfig()
		cfg.MaxRounds = 1
		assert.NoError(t, cfg.Validate())
	})
}
