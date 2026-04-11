//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultWarConfig(t *testing.T) {
	c := DefaultWarConfig()
	assert.Equal(t, WarDefaultMaxRounds, c.MaxRounds)
	assert.NoError(t, c.Validate())
}

func TestWarConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     WarConfig
		wantErr bool
	}{
		{"min boundary", WarConfig{MaxRounds: WarMinMaxRounds}, false},
		{"max boundary", WarConfig{MaxRounds: WarMaxMaxRounds}, false},
		{"below min", WarConfig{MaxRounds: WarMinMaxRounds - 1}, true},
		{"above max", WarConfig{MaxRounds: WarMaxMaxRounds + 1}, true},
		{"default", DefaultWarConfig(), false},
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

func TestWarConfig_JSON(t *testing.T) {
	src := WarConfig{MaxRounds: 250}
	data, err := json.Marshal(src)
	assert.NoError(t, err)

	var dst WarConfig
	assert.NoError(t, json.Unmarshal(data, &dst))
	assert.Equal(t, src, dst)
}
