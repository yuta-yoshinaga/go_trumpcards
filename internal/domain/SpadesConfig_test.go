//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultSpadesConfig(t *testing.T) {
	cfg := domain.DefaultSpadesConfig()
	assert.Equal(t, domain.SpadesCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 500, cfg.PointLimit)
	assert.Equal(t, 100, cfg.NilBonus)
	assert.Equal(t, 10, cfg.BagPenaltyThreshold)
}

func TestSpadesConfig_Validate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid easy difficulty", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.CpuDifficulty = domain.SpadesCpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid hard difficulty", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.CpuDifficulty = domain.SpadesCpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})

	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.CpuDifficulty = domain.SpadesCpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.CpuDifficulty = domain.SpadesCpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("point limit zero", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.PointLimit = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit negative", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.PointLimit = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit 1 is valid", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.PointLimit = 1
		assert.NoError(t, cfg.Validate())
	})

	t.Run("nil bonus negative", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.NilBonus = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil bonus")
	})

	t.Run("nil bonus zero is valid", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.NilBonus = 0
		assert.NoError(t, cfg.Validate())
	})

	t.Run("bag penalty threshold zero", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.BagPenaltyThreshold = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bag penalty threshold")
	})

	t.Run("bag penalty threshold 1 is valid", func(t *testing.T) {
		cfg := domain.DefaultSpadesConfig()
		cfg.BagPenaltyThreshold = 1
		assert.NoError(t, cfg.Validate())
	})
}
