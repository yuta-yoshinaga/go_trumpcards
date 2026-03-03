package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultDoubtConfig(t *testing.T) {
	cfg := domain.DefaultDoubtConfig()
	assert.Equal(t, 10, cfg.DoubtWindowSec)
	assert.Equal(t, domain.DoubtMemoryLevelNormal, cfg.CpuMemoryLevel)
}

func TestDoubtMemoryLevelConstants(t *testing.T) {
	assert.Equal(t, domain.DoubtMemoryLevel(0), domain.DoubtMemoryLevelEasy)
	assert.Equal(t, domain.DoubtMemoryLevel(1), domain.DoubtMemoryLevelNormal)
	assert.Equal(t, domain.DoubtMemoryLevel(2), domain.DoubtMemoryLevelHard)
}
