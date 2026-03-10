package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultMemoryConfig(t *testing.T) {
	cfg := DefaultMemoryConfig()
	assert.Equal(t, MemoryCpuDifficultyNormal, cfg.CpuDifficulty)
}

func TestMemoryCpuDifficultyConstants(t *testing.T) {
	assert.Equal(t, MemoryCpuDifficulty(0), MemoryCpuDifficultyEasy)
	assert.Equal(t, MemoryCpuDifficulty(1), MemoryCpuDifficultyNormal)
	assert.Equal(t, MemoryCpuDifficulty(2), MemoryCpuDifficultyHard)
}
