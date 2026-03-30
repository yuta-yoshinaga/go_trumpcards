//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultBridgeConfig(t *testing.T) {
	cfg := DefaultBridgeConfig()
	assert.Equal(t, BridgeCpuDifficultyNormal, cfg.CpuDifficulty)
}

func TestBridgeConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  BridgeConfig
		wantErr bool
	}{
		{"valid default", DefaultBridgeConfig(), false},
		{"valid easy", BridgeConfig{CpuDifficulty: BridgeCpuDifficultyEasy}, false},
		{"valid hard", BridgeConfig{CpuDifficulty: BridgeCpuDifficultyHard}, false},
		{"invalid difficulty low", BridgeConfig{CpuDifficulty: -1}, true},
		{"invalid difficulty high", BridgeConfig{CpuDifficulty: 3}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
