//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultTarneebConfig(t *testing.T) {
	cfg := domain.DefaultTarneebConfig()
	assert.Equal(t, domain.TarneebCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, domain.TarneebDefaultPointLimit, cfg.PointLimit)
	assert.Equal(t, domain.TarneebDefaultMinBid, cfg.MinBid)
}

func TestTarneebConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.TarneebConfig
		wantErr bool
	}{
		{"default", domain.DefaultTarneebConfig(), false},
		{"easy", domain.TarneebConfig{CpuDifficulty: domain.TarneebCpuDifficultyEasy, PointLimit: 1, MinBid: 1}, false},
		{"hard high limit", domain.TarneebConfig{CpuDifficulty: domain.TarneebCpuDifficultyHard, PointLimit: domain.TarneebMaxPointLimit, MinBid: 13}, false},
		{"bad difficulty", domain.TarneebConfig{CpuDifficulty: -1, PointLimit: 31, MinBid: 7}, true},
		{"bad difficulty high", domain.TarneebConfig{CpuDifficulty: domain.TarneebCpuDifficultyHard + 1, PointLimit: 31, MinBid: 7}, true},
		{"point limit zero", domain.TarneebConfig{CpuDifficulty: domain.TarneebCpuDifficultyNormal, PointLimit: 0, MinBid: 7}, true},
		{"point limit too high", domain.TarneebConfig{CpuDifficulty: domain.TarneebCpuDifficultyNormal, PointLimit: domain.TarneebMaxPointLimit + 1, MinBid: 7}, true},
		{"min bid zero", domain.TarneebConfig{CpuDifficulty: domain.TarneebCpuDifficultyNormal, PointLimit: 31, MinBid: 0}, true},
		{"min bid too high", domain.TarneebConfig{CpuDifficulty: domain.TarneebCpuDifficultyNormal, PointLimit: 31, MinBid: 14}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
