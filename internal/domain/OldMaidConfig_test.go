package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultOldMaidConfig(t *testing.T) {
	cfg := domain.DefaultOldMaidConfig()
	assert.Equal(t, domain.OldMaidModeNormal, cfg.Mode)
	assert.False(t, cfg.CpuPlacementStrategy)
}

func TestOldMaidMode_Values(t *testing.T) {
	assert.Equal(t, domain.OldMaidMode(0), domain.OldMaidModeNormal)
	assert.Equal(t, domain.OldMaidMode(1), domain.OldMaidModeJijiNuki)
}
