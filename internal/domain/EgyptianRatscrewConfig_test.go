//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEgyptianRatscrewConfig_Default(t *testing.T) {
	c := DefaultEgyptianRatscrewConfig()
	assert.Equal(t, EgyptianRatscrewCpuNormal, c.CpuDifficulty)
}

func TestEgyptianRatscrewConfig_Validate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     EgyptianRatscrewConfig
		wantErr bool
	}{
		{"easy", EgyptianRatscrewConfig{CpuDifficulty: EgyptianRatscrewCpuEasy}, false},
		{"normal", EgyptianRatscrewConfig{CpuDifficulty: EgyptianRatscrewCpuNormal}, false},
		{"hard", EgyptianRatscrewConfig{CpuDifficulty: EgyptianRatscrewCpuHard}, false},
		{"too low", EgyptianRatscrewConfig{CpuDifficulty: -1}, true},
		{"too high", EgyptianRatscrewConfig{CpuDifficulty: 99}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEgyptianRatscrewConfig_ReactionMeanMs(t *testing.T) {
	cases := []struct {
		diff EgyptianRatscrewCpuDifficulty
		want int
	}{
		{EgyptianRatscrewCpuEasy, 1100},
		{EgyptianRatscrewCpuNormal, 600},
		{EgyptianRatscrewCpuHard, 300},
		{99, 600}, // fallback to normal
	}
	for _, tc := range cases {
		c := EgyptianRatscrewConfig{CpuDifficulty: tc.diff}
		assert.Equal(t, tc.want, c.ReactionMeanMs())
	}
}

func TestEgyptianRatscrewConfig_ReactionStdDevMs(t *testing.T) {
	cases := []struct {
		diff EgyptianRatscrewCpuDifficulty
		want int
	}{
		{EgyptianRatscrewCpuEasy, 300},
		{EgyptianRatscrewCpuNormal, 200},
		{EgyptianRatscrewCpuHard, 120},
		{99, 200},
	}
	for _, tc := range cases {
		c := EgyptianRatscrewConfig{CpuDifficulty: tc.diff}
		assert.Equal(t, tc.want, c.ReactionStdDevMs())
	}
}

func TestEgyptianRatscrewConfig_JSONRoundtrip(t *testing.T) {
	c := EgyptianRatscrewConfig{CpuDifficulty: EgyptianRatscrewCpuHard}
	data, err := json.Marshal(c)
	assert.NoError(t, err)
	var d EgyptianRatscrewConfig
	assert.NoError(t, json.Unmarshal(data, &d))
	assert.Equal(t, c, d)
}

func TestEgyptianRatscrewConfig_UnmarshalInvalid(t *testing.T) {
	var c EgyptianRatscrewConfig
	assert.Error(t, c.UnmarshalJSON([]byte("not json")))
}
