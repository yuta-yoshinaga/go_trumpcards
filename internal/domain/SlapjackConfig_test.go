//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlapjackConfig_Default(t *testing.T) {
	cfg := DefaultSlapjackConfig()
	assert.Equal(t, SlapjackCpuNormal, cfg.CpuDifficulty)
	assert.NoError(t, cfg.Validate())
}

func TestSlapjackConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SlapjackConfig
		wantErr bool
	}{
		{"easy ok", SlapjackConfig{CpuDifficulty: SlapjackCpuEasy}, false},
		{"normal ok", SlapjackConfig{CpuDifficulty: SlapjackCpuNormal}, false},
		{"hard ok", SlapjackConfig{CpuDifficulty: SlapjackCpuHard}, false},
		{"too low", SlapjackConfig{CpuDifficulty: -1}, true},
		{"too high", SlapjackConfig{CpuDifficulty: 99}, true},
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

func TestSlapjackConfig_ReactionParameters(t *testing.T) {
	tests := []struct {
		diff       SlapjackCpuDifficulty
		wantMean   int
		wantStdDev int
	}{
		{SlapjackCpuEasy, 1100, 300},
		{SlapjackCpuNormal, 600, 200},
		{SlapjackCpuHard, 300, 120},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			c := SlapjackConfig{CpuDifficulty: tt.diff}
			assert.Equal(t, tt.wantMean, c.ReactionMeanMs())
			assert.Equal(t, tt.wantStdDev, c.ReactionStdDevMs())
		})
	}
}

func TestSlapjackConfig_ReactionFallbackToNormal(t *testing.T) {
	// 未知の難易度は Normal の値を返す (defensive default branch)
	c := SlapjackConfig{CpuDifficulty: SlapjackCpuDifficulty(99)}
	assert.Equal(t, 600, c.ReactionMeanMs())
	assert.Equal(t, 200, c.ReactionStdDevMs())
}

func TestSlapjackConfig_JSON(t *testing.T) {
	original := SlapjackConfig{CpuDifficulty: SlapjackCpuHard}
	data, err := json.Marshal(original)
	assert.NoError(t, err)

	var decoded SlapjackConfig
	assert.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, decoded)
}
