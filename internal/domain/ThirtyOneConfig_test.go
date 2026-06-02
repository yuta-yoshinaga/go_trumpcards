package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultThirtyOneConfig(t *testing.T) {
	cfg := DefaultThirtyOneConfig()
	assert.Equal(t, ThirtyOneCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 3, cfg.InitialLives)
	assert.NoError(t, cfg.Validate())
}

func TestThirtyOneConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ThirtyOneConfig
		wantErr bool
	}{
		{"valid", ThirtyOneConfig{CpuDifficulty: ThirtyOneCpuDifficultyHard, InitialLives: 5}, false},
		{"bad difficulty low", ThirtyOneConfig{CpuDifficulty: -1, InitialLives: 3}, true},
		{"bad difficulty high", ThirtyOneConfig{CpuDifficulty: 9, InitialLives: 3}, true},
		{"lives too low", ThirtyOneConfig{CpuDifficulty: ThirtyOneCpuDifficultyNormal, InitialLives: 0}, true},
		{"lives too high", ThirtyOneConfig{CpuDifficulty: ThirtyOneCpuDifficultyNormal, InitialLives: 99}, true},
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
