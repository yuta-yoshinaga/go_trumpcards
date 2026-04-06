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
}

func TestPigsTailConfig_Validate(t *testing.T) {
	cfg := PigsTailConfig{CpuHesitationEnabled: true}
	assert.NoError(t, cfg.Validate())
}

func TestPigsTailConfig_JSONRoundTrip(t *testing.T) {
	cfg := PigsTailConfig{CpuHesitationEnabled: true}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var restored PigsTailConfig
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)
	assert.Equal(t, cfg, restored)
}

func TestPigsTailConfig_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var cfg PigsTailConfig
	err := json.Unmarshal([]byte(`{invalid`), &cfg)
	assert.Error(t, err)
}
