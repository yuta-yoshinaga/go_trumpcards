//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPigsTailConfig(t *testing.T) {
	cfg := DefaultPigsTailConfig()
	assert.False(t, cfg.CpuHesitationEnabled)
	assert.Equal(t, PigsTailPlayerCnt, cfg.PlayerCount)
}

func TestPigsTailConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		wantErr bool
	}{
		{"min valid", PigsTailMinPlayers, false},
		{"default valid", PigsTailPlayerCnt, false},
		{"max valid", PigsTailMaxPlayers, false},
		{"below min", PigsTailMinPlayers - 1, true},
		{"above max", PigsTailMaxPlayers + 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := PigsTailConfig{CpuHesitationEnabled: true, PlayerCount: tt.count}
			if tt.wantErr {
				assert.Error(t, cfg.Validate())
			} else {
				assert.NoError(t, cfg.Validate())
			}
		})
	}
}

func TestPigsTailConfig_JSONRoundTrip(t *testing.T) {
	cfg := PigsTailConfig{CpuHesitationEnabled: true, PlayerCount: 3}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var restored PigsTailConfig
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)
	assert.Equal(t, cfg, restored)
}

func TestPigsTailConfig_UnmarshalJSON_LegacyDefaultsPlayerCount(t *testing.T) {
	// 旧スナップショット (pc フィールドなし) は既定人数へ丸める。
	var cfg PigsTailConfig
	err := json.Unmarshal([]byte(`{"ch":true}`), &cfg)
	require.NoError(t, err)
	assert.Equal(t, PigsTailPlayerCnt, cfg.PlayerCount)
}

func TestPigsTailConfig_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var cfg PigsTailConfig
	err := json.Unmarshal([]byte(`{invalid`), &cfg)
	assert.Error(t, err)
}
