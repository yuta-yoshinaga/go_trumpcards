package domain

import (
	"encoding/json"
	"testing"
)

func TestTrucoConfigUnmarshalLegacyDifficulty(t *testing.T) {
	var cfg TrucoConfig
	if err := json.Unmarshal([]byte(`{"cd":0,"mt":15}`), &cfg); err != nil {
		t.Fatalf("Unmarshal legacy config: %v", err)
	}
	if cfg.MatchTarget != 15 {
		t.Errorf("MatchTarget = %d, want 15", cfg.MatchTarget)
	}
}

func TestDefaultTrucoConfig(t *testing.T) {
	c := DefaultTrucoConfig()
	if c.MatchTarget != TrucoDefaultMatchTarget {
		t.Errorf("MatchTarget = %d, want %d", c.MatchTarget, TrucoDefaultMatchTarget)
	}
}

func TestTrucoConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     TrucoConfig
		wantErr bool
	}{
		{"default", DefaultTrucoConfig(), false},
		{"min target", TrucoConfig{MatchTarget: TrucoMinMatchTarget}, false},
		{"max target", TrucoConfig{MatchTarget: TrucoMaxMatchTarget}, false},
		{"target too low", TrucoConfig{MatchTarget: 0}, true},
		{"target too high", TrucoConfig{MatchTarget: TrucoMaxMatchTarget + 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTrucoConfigNormalized(t *testing.T) {
	if got := (TrucoConfig{MatchTarget: 0}).normalized().MatchTarget; got != TrucoDefaultMatchTarget {
		t.Errorf("normalized 0 -> %d, want %d", got, TrucoDefaultMatchTarget)
	}
	if got := (TrucoConfig{MatchTarget: 999}).normalized().MatchTarget; got != TrucoDefaultMatchTarget {
		t.Errorf("normalized 999 -> %d, want %d", got, TrucoDefaultMatchTarget)
	}
	if got := (TrucoConfig{MatchTarget: 30}).normalized().MatchTarget; got != 30 {
		t.Errorf("normalized 30 -> %d, want 30", got)
	}
}
