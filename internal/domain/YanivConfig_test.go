package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultYanivConfig(t *testing.T) {
	cfg := DefaultYanivConfig()
	assert.Equal(t, YanivCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 200, cfg.ScoreLimit)
	require.NoError(t, cfg.Validate())
}

func TestYanivConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     YanivConfig
		wantErr bool
	}{
		{"valid", YanivConfig{CpuDifficulty: YanivCpuDifficultyHard, ScoreLimit: 100}, false},
		{"difficulty too low", YanivConfig{CpuDifficulty: -1, ScoreLimit: 200}, true},
		{"difficulty too high", YanivConfig{CpuDifficulty: 99, ScoreLimit: 200}, true},
		{"limit too low", YanivConfig{CpuDifficulty: YanivCpuDifficultyNormal, ScoreLimit: YanivMinScoreLimit - 1}, true},
		{"limit too high", YanivConfig{CpuDifficulty: YanivCpuDifficultyNormal, ScoreLimit: YanivMaxScoreLimit + 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Error(t, tt.cfg.Validate())
			} else {
				assert.NoError(t, tt.cfg.Validate())
			}
		})
	}
}
