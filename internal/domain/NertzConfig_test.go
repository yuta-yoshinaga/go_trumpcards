//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDefaultNertzConfig(t *testing.T) {
	cfg := domain.DefaultNertzConfig()
	assert.Equal(t, domain.NertzPlayerCntDefault, cfg.PlayerCount)
	assert.Equal(t, 3, cfg.DrawCount)
	assert.Equal(t, domain.NertzTargetScoreDefault, cfg.TargetScore)
	assert.Equal(t, domain.NertzCpuDifficultyNormal, cfg.CpuDifficulty)
}

func TestNertzCpuDifficultyConstants(t *testing.T) {
	assert.Equal(t, domain.NertzCpuDifficulty(0), domain.NertzCpuDifficultyEasy)
	assert.Equal(t, domain.NertzCpuDifficulty(1), domain.NertzCpuDifficultyNormal)
	assert.Equal(t, domain.NertzCpuDifficulty(2), domain.NertzCpuDifficultyHard)
}

func TestNertzConfig_Validate(t *testing.T) {
	base := domain.DefaultNertzConfig()
	tests := []struct {
		name    string
		mutate  func(*domain.NertzConfig)
		wantErr bool
	}{
		{"default", func(c *domain.NertzConfig) {}, false},
		{"min players", func(c *domain.NertzConfig) { c.PlayerCount = domain.NertzPlayerCntMin }, false},
		{"max players", func(c *domain.NertzConfig) { c.PlayerCount = domain.NertzPlayerCntMax }, false},
		{"too few players", func(c *domain.NertzConfig) { c.PlayerCount = domain.NertzPlayerCntMin - 1 }, true},
		{"too many players", func(c *domain.NertzConfig) { c.PlayerCount = domain.NertzPlayerCntMax + 1 }, true},
		{"draw count 1", func(c *domain.NertzConfig) { c.DrawCount = 1 }, false},
		{"draw count 3", func(c *domain.NertzConfig) { c.DrawCount = 3 }, false},
		{"draw count 2 invalid", func(c *domain.NertzConfig) { c.DrawCount = 2 }, true},
		{"draw count 0 invalid", func(c *domain.NertzConfig) { c.DrawCount = 0 }, true},
		{"min target score", func(c *domain.NertzConfig) { c.TargetScore = domain.NertzTargetScoreMin }, false},
		{"max target score", func(c *domain.NertzConfig) { c.TargetScore = domain.NertzTargetScoreMax }, false},
		{"target score too low", func(c *domain.NertzConfig) { c.TargetScore = domain.NertzTargetScoreMin - 1 }, true},
		{"target score too high", func(c *domain.NertzConfig) { c.TargetScore = domain.NertzTargetScoreMax + 1 }, true},
		{"cpu difficulty too low", func(c *domain.NertzConfig) { c.CpuDifficulty = -1 }, true},
		{"cpu difficulty too high", func(c *domain.NertzConfig) { c.CpuDifficulty = 3 }, true},
		{"cpu tick moves negative", func(c *domain.NertzConfig) { c.CpuTickMoves = -1 }, true},
		{"cpu tick moves zero is auto", func(c *domain.NertzConfig) { c.CpuTickMoves = 0 }, false},
		{"cpu tick moves max", func(c *domain.NertzConfig) { c.CpuTickMoves = domain.NertzCpuTickMovesMax }, false},
		{"cpu tick moves over max", func(c *domain.NertzConfig) { c.CpuTickMoves = domain.NertzCpuTickMovesMax + 1 }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := base
			tt.mutate(&cfg)
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNertzConfig_JSON(t *testing.T) {
	cfg := domain.NertzConfig{
		PlayerCount:   3,
		DrawCount:     1,
		TargetScore:   75,
		CpuDifficulty: domain.NertzCpuDifficultyHard,
		CpuTickMoves:  4,
	}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var restored domain.NertzConfig
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)
	assert.Equal(t, cfg, restored)
}

func TestNertzConfig_ResolvedCpuTickMoves(t *testing.T) {
	tests := []struct {
		name string
		cfg  domain.NertzConfig
		want int
	}{
		{"explicit override", domain.NertzConfig{CpuTickMoves: 7}, 7},
		{"auto easy", domain.NertzConfig{CpuDifficulty: domain.NertzCpuDifficultyEasy}, 1},
		{"auto normal", domain.NertzConfig{CpuDifficulty: domain.NertzCpuDifficultyNormal}, 3},
		{"auto hard", domain.NertzConfig{CpuDifficulty: domain.NertzCpuDifficultyHard}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cfg.ResolvedCpuTickMoves())
		})
	}
}
