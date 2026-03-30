//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultSpeedConfig(t *testing.T) {
	cfg := domain.DefaultSpeedConfig()
	assert.Equal(t, domain.SpeedCpuDifficultyNormal, cfg.CpuDifficulty)
}

func TestSpeedCpuDifficultyConstants(t *testing.T) {
	assert.Equal(t, domain.SpeedCpuDifficulty(0), domain.SpeedCpuDifficultyEasy)
	assert.Equal(t, domain.SpeedCpuDifficulty(1), domain.SpeedCpuDifficultyNormal)
	assert.Equal(t, domain.SpeedCpuDifficulty(2), domain.SpeedCpuDifficultyHard)
}

func TestSpeedConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.SpeedConfig
		wantErr bool
	}{
		{"easy", domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyEasy}, false},
		{"normal", domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyNormal}, false},
		{"hard", domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyHard}, false},
		{"too low", domain.SpeedConfig{CpuDifficulty: -1}, true},
		{"too high", domain.SpeedConfig{CpuDifficulty: 3}, true},
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

func TestSpeedConfig_JSON(t *testing.T) {
	cfg := domain.SpeedConfig{CpuDifficulty: domain.SpeedCpuDifficultyHard}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var restored domain.SpeedConfig
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)
	assert.Equal(t, cfg, restored)
}
