//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoFishConfig_Default(t *testing.T) {
	cfg := DefaultGoFishConfig()
	assert.Equal(t, GoFishCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.False(t, cfg.CpuMetaAI)
}

func TestGoFishConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     GoFishConfig
		wantErr bool
	}{
		{"easy", GoFishConfig{CpuDifficulty: GoFishCpuDifficultyEasy}, false},
		{"normal", GoFishConfig{CpuDifficulty: GoFishCpuDifficultyNormal}, false},
		{"hard", GoFishConfig{CpuDifficulty: GoFishCpuDifficultyHard}, false},
		{"negative", GoFishConfig{CpuDifficulty: GoFishCpuDifficulty(-1)}, true},
		{"too high", GoFishConfig{CpuDifficulty: GoFishCpuDifficulty(3)}, true},
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

func TestGoFishConfig_JSON(t *testing.T) {
	cfg := GoFishConfig{CpuDifficulty: GoFishCpuDifficultyHard, CpuMetaAI: true}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var restored GoFishConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, cfg, restored)
}
