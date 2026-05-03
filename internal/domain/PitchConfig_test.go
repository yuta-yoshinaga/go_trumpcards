//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultPitchConfig(t *testing.T) {
	cfg := domain.DefaultPitchConfig()
	assert.Equal(t, domain.PitchCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 7, cfg.PointLimit)
	assert.NoError(t, cfg.Validate())
}

func TestPitchConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.PitchConfig
		wantErr bool
	}{
		{"default ok", domain.DefaultPitchConfig(), false},
		{"hard difficulty ok", domain.PitchConfig{CpuDifficulty: domain.PitchCpuDifficultyHard, PointLimit: 11}, false},
		{"negative difficulty rejected", domain.PitchConfig{CpuDifficulty: domain.PitchCpuDifficulty(-1), PointLimit: 7}, true},
		{"out-of-range difficulty rejected", domain.PitchConfig{CpuDifficulty: domain.PitchCpuDifficulty(99), PointLimit: 7}, true},
		{"zero point limit rejected", domain.PitchConfig{CpuDifficulty: domain.PitchCpuDifficultyNormal, PointLimit: 0}, true},
		{"negative point limit rejected", domain.PitchConfig{CpuDifficulty: domain.PitchCpuDifficultyNormal, PointLimit: -1}, true},
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
