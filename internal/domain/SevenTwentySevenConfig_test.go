//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSevenTwentySevenConfig_Default(t *testing.T) {
	cfg := DefaultSevenTwentySevenConfig()
	assert.NoError(t, cfg.Validate(), "既定値が自分の検証を通らない")
	assert.GreaterOrEqual(t, cfg.PlayerCount, SevenTwentySevenMinPlayerCount)
	assert.LessOrEqual(t, cfg.PlayerCount, SevenTwentySevenMaxPlayerCount)
}

// **境界の両側を見る。** 片側だけだと「常にエラー」でも通る。
func TestSevenTwentySevenConfig_Validate(t *testing.T) {
	for _, tt := range []struct {
		name    string
		mutate  func(*SevenTwentySevenConfig)
		wantErr bool
	}{
		{"players at min", func(c *SevenTwentySevenConfig) { c.PlayerCount = SevenTwentySevenMinPlayerCount }, false},
		{"players below min", func(c *SevenTwentySevenConfig) { c.PlayerCount = SevenTwentySevenMinPlayerCount - 1 }, true},
		{"players at max", func(c *SevenTwentySevenConfig) { c.PlayerCount = SevenTwentySevenMaxPlayerCount }, false},
		{"players above max", func(c *SevenTwentySevenConfig) { c.PlayerCount = SevenTwentySevenMaxPlayerCount + 1 }, true},
		{"ante at min", func(c *SevenTwentySevenConfig) { c.Ante = SevenTwentySevenMinAnte }, false},
		{"ante below min", func(c *SevenTwentySevenConfig) { c.Ante = SevenTwentySevenMinAnte - 1 }, true},
		{"ante above max", func(c *SevenTwentySevenConfig) { c.Ante = SevenTwentySevenMaxAnte + 1 }, true},
		{"chips at min", func(c *SevenTwentySevenConfig) { c.StartingChips = SevenTwentySevenMinStartingChips }, false},
		{"chips below min", func(c *SevenTwentySevenConfig) { c.StartingChips = SevenTwentySevenMinStartingChips - 1 }, true},
		{"rounds at min", func(c *SevenTwentySevenConfig) { c.TargetRounds = SevenTwentySevenMinTargetRounds }, false},
		{"rounds below min", func(c *SevenTwentySevenConfig) { c.TargetRounds = SevenTwentySevenMinTargetRounds - 1 }, true},
		{"rounds above max", func(c *SevenTwentySevenConfig) { c.TargetRounds = SevenTwentySevenMaxTargetRounds + 1 }, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultSevenTwentySevenConfig()
			tt.mutate(&cfg)
			if tt.wantErr {
				assert.Error(t, cfg.Validate())
				return
			}
			assert.NoError(t, cfg.Validate())
		})
	}
}

// **アンティより少ない初期チップは弾く。** 1 ラウンドも打てない卓ができてしまう。
func TestSevenTwentySevenConfig_StartingChipsMustCoverTheAnte(t *testing.T) {
	cfg := DefaultSevenTwentySevenConfig()
	cfg.Ante = 50
	cfg.StartingChips = 40
	assert.Error(t, cfg.Validate())

	cfg.StartingChips = 50 // ちょうどなら通る
	assert.NoError(t, cfg.Validate())
}

func TestSevenTwentySevenConfig_JSONRoundTrip(t *testing.T) {
	cfg := DefaultSevenTwentySevenConfig()
	cfg.PlayerCount = 5
	cfg.Ante = 25
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var got SevenTwentySevenConfig
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, cfg, got)
}
