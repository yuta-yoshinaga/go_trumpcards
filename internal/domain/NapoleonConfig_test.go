//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultNapoleonConfig(t *testing.T) {
	cfg := domain.DefaultNapoleonConfig()
	assert.Equal(t, domain.NapoleonCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 12, cfg.MinBid)
	assert.Equal(t, 100, cfg.PointLimit)
}

func TestNapoleonConfig_Validate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid easy difficulty", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.CpuDifficulty = domain.NapoleonCpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid hard difficulty", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.CpuDifficulty = domain.NapoleonCpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})

	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.CpuDifficulty = domain.NapoleonCpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.CpuDifficulty = domain.NapoleonCpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("min bid zero", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.MinBid = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min bid")
	})

	t.Run("min bid too high", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.MinBid = domain.NapoleonMaxPictureCards + 1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min bid")
	})

	t.Run("min bid 1 is valid", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.MinBid = 1
		assert.NoError(t, cfg.Validate())
	})

	t.Run("min bid max is valid", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.MinBid = domain.NapoleonMaxPictureCards
		assert.NoError(t, cfg.Validate())
	})

	t.Run("point limit zero", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.PointLimit = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit negative", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.PointLimit = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit 1 is valid", func(t *testing.T) {
		cfg := domain.DefaultNapoleonConfig()
		cfg.PointLimit = 1
		assert.NoError(t, cfg.Validate())
	})
}
