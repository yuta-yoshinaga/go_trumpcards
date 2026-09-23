//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultBatakConfig(t *testing.T) {
	cfg := domain.DefaultBatakConfig()
	assert.Equal(t, domain.BatakCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, domain.BatakDefaultMaxRounds, cfg.MaxRounds)
	assert.Equal(t, 5, domain.BatakMinBid)
	assert.Equal(t, 13, domain.BatakMaxBid)
	assert.Equal(t, 0, domain.BatakPassBid)
}

func TestBatakConfig_Validate(t *testing.T) {
	t.Run("valid default", func(t *testing.T) {
		assert.NoError(t, domain.DefaultBatakConfig().Validate())
	})

	t.Run("easy difficulty valid", func(t *testing.T) {
		cfg := domain.DefaultBatakConfig()
		cfg.CpuDifficulty = domain.BatakCpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})

	t.Run("hard difficulty valid", func(t *testing.T) {
		cfg := domain.DefaultBatakConfig()
		cfg.CpuDifficulty = domain.BatakCpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})

	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultBatakConfig()
		cfg.CpuDifficulty = domain.BatakCpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultBatakConfig()
		cfg.CpuDifficulty = domain.BatakCpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("max rounds zero invalid", func(t *testing.T) {
		cfg := domain.DefaultBatakConfig()
		cfg.MaxRounds = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max rounds")
	})

	t.Run("max rounds negative", func(t *testing.T) {
		cfg := domain.DefaultBatakConfig()
		cfg.MaxRounds = -3
		err := cfg.Validate()
		assert.Error(t, err)
	})

	t.Run("max rounds 1 valid", func(t *testing.T) {
		cfg := domain.DefaultBatakConfig()
		cfg.MaxRounds = 1
		assert.NoError(t, cfg.Validate())
	})

	t.Run("max rounds at cap valid", func(t *testing.T) {
		cfg := domain.DefaultBatakConfig()
		cfg.MaxRounds = domain.BatakMaxAllowedRounds
		assert.NoError(t, cfg.Validate())
	})

	t.Run("max rounds above cap invalid", func(t *testing.T) {
		cfg := domain.DefaultBatakConfig()
		cfg.MaxRounds = domain.BatakMaxAllowedRounds + 1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max rounds")
	})
}
