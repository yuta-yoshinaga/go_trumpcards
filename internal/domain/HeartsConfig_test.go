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
