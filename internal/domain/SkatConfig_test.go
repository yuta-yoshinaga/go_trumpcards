//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func TestSkatConfigDefault(t *testing.T) {
	c := DefaultSkatConfig()
	if c.CpuDifficulty != SkatCpuDifficultyNormal {
		t.Fatalf("default difficulty = %v, want Normal", c.CpuDifficulty)
	}
	if c.TargetScore != 500 {
		t.Fatalf("default target = %d, want 500", c.TargetScore)
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
}

func TestSkatConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SkatConfig
		wantErr bool
	}{
		{"ok", SkatConfig{CpuDifficulty: SkatCpuDifficultyEasy, TargetScore: 100}, false},
		{"bad difficulty", SkatConfig{CpuDifficulty: SkatCpuDifficulty(99), TargetScore: 100}, true},
		{"bad target", SkatConfig{CpuDifficulty: SkatCpuDifficultyNormal, TargetScore: 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestSkatConfigJSONRoundTrip(t *testing.T) {
	in := DefaultSkatConfig()
	in.CpuDifficulty = SkatCpuDifficultyHard
	in.TargetScore = 250
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out SkatConfig
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out != in {
		t.Fatalf("roundtrip mismatch: %+v vs %+v", out, in)
	}
}
