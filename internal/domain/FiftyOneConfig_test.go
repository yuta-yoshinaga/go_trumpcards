//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFiftyOneConfig(t *testing.T) {
	c := DefaultFiftyOneConfig()
	assert.Equal(t, FiftyOneDifficultyNormal, c.CpuDifficulty)
	assert.NoError(t, c.Validate())
}

func TestFiftyOneConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     FiftyOneConfig
		wantErr bool
	}{
		{"easy", FiftyOneConfig{CpuDifficulty: FiftyOneDifficultyEasy}, false},
		{"normal", FiftyOneConfig{CpuDifficulty: FiftyOneDifficultyNormal}, false},
		{"hard", FiftyOneConfig{CpuDifficulty: FiftyOneDifficultyHard}, false},
		{"below min", FiftyOneConfig{CpuDifficulty: FiftyOneCpuDifficulty(-1)}, true},
		{"above max", FiftyOneConfig{CpuDifficulty: FiftyOneCpuDifficulty(3)}, true},
		{"default", DefaultFiftyOneConfig(), false},
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

func TestFiftyOneConfig_JSON(t *testing.T) {
	src := FiftyOneConfig{CpuDifficulty: FiftyOneDifficultyHard}
	data, err := json.Marshal(src)
	assert.NoError(t, err)

	var dst FiftyOneConfig
	assert.NoError(t, json.Unmarshal(data, &dst))
	assert.Equal(t, src, dst)
}
