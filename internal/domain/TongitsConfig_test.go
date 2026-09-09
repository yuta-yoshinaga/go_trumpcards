//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultTongitsConfig(t *testing.T) {
	cfg := domain.DefaultTongitsConfig()

	assert.Equal(t, domain.TongitsCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 50, cfg.PointLimit)
}

func TestTongitsConfig_Validate(t *testing.T) {
	tests := []struct {
		name          string
		cpuDifficulty domain.TongitsCpuDifficulty
		pointLimit    int
		wantError     string
	}{
		{name: "default", cpuDifficulty: domain.TongitsCpuDifficultyNormal, pointLimit: 50},
		{name: "minimum difficulty", cpuDifficulty: domain.TongitsCpuDifficultyEasy, pointLimit: 1},
		{name: "maximum difficulty", cpuDifficulty: domain.TongitsCpuDifficultyHard, pointLimit: 1},
		{name: "difficulty below range", cpuDifficulty: -1, pointLimit: 50, wantError: "CPU difficulty"},
		{name: "difficulty above range", cpuDifficulty: 3, pointLimit: 50, wantError: "CPU difficulty"},
		{name: "point limit below minimum", cpuDifficulty: domain.TongitsCpuDifficultyNormal, pointLimit: 0, wantError: "point limit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (domain.TongitsConfig{
				CpuDifficulty: tt.cpuDifficulty,
				PointLimit:    tt.pointLimit,
			}).Validate()
			if tt.wantError == "" {
				assert.NoError(t, err)
				return
			}
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), tt.wantError)
			}
		})
	}
}
