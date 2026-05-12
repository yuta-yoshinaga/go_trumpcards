//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultMightyConfig(t *testing.T) {
	cfg := domain.DefaultMightyConfig()
	assert.Equal(t, domain.MightyCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 13, cfg.MinBid)
	assert.Equal(t, 2, cfg.NoTrumpExtra)
	assert.Equal(t, 100, cfg.PointLimit)
}

func TestMightyConfig_Validate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid easy difficulty", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.CpuDifficulty = domain.MightyCpuDifficultyEasy
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid hard difficulty", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.CpuDifficulty = domain.MightyCpuDifficultyHard
		assert.NoError(t, cfg.Validate())
	})

	t.Run("difficulty below range", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.CpuDifficulty = domain.MightyCpuDifficulty(-1)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("difficulty above range", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.CpuDifficulty = domain.MightyCpuDifficulty(3)
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CPU difficulty")
	})

	t.Run("min bid zero", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.MinBid = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min bid")
	})

	t.Run("min bid too high", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.MinBid = domain.MightyMaxPoints + 1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min bid")
	})

	t.Run("min bid 1 is valid", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.MinBid = 1
		assert.NoError(t, cfg.Validate())
	})

	t.Run("min bid max is valid", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.MinBid = domain.MightyMaxPoints
		assert.NoError(t, cfg.Validate())
	})

	t.Run("no trump extra negative", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.NoTrumpExtra = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no trump extra")
	})

	t.Run("no trump extra zero is valid", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.NoTrumpExtra = 0
		assert.NoError(t, cfg.Validate())
	})

	t.Run("no trump extra too high", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.NoTrumpExtra = domain.MightyMaxPoints
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no trump extra")
	})

	t.Run("point limit zero", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.PointLimit = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit negative", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.PointLimit = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "point limit")
	})

	t.Run("point limit 1 is valid", func(t *testing.T) {
		cfg := domain.DefaultMightyConfig()
		cfg.PointLimit = 1
		assert.NoError(t, cfg.Validate())
	})
}

func TestMightyConfig_JSONRoundTrip(t *testing.T) {
	orig := domain.MightyConfig{
		CpuDifficulty: domain.MightyCpuDifficultyHard,
		MinBid:        15,
		NoTrumpExtra:  3,
		PointLimit:    150,
	}
	data, err := json.Marshal(orig)
	assert.NoError(t, err)

	var got domain.MightyConfig
	assert.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, orig, got)
}
