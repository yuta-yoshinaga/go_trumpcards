//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"testing"
)

func TestDefaultPutConfig(t *testing.T) {
	c := DefaultPutConfig()
	if c.MatchTarget != PutDefaultMatchTarget {
		t.Errorf("MatchTarget = %d, want %d", c.MatchTarget, PutDefaultMatchTarget)
	}
}

func TestPutConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     PutConfig
		wantErr bool
	}{
		{"default", DefaultPutConfig(), false},
		{"min target", PutConfig{MatchTarget: PutMinMatchTarget}, false},
		{"max target", PutConfig{MatchTarget: PutMaxMatchTarget}, false},
		{"target too low", PutConfig{MatchTarget: 0}, true},
		{"target too high", PutConfig{MatchTarget: PutMaxMatchTarget + 1}, true},
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

func TestPutConfigNormalized(t *testing.T) {
	if got := (PutConfig{MatchTarget: 0}).normalized().MatchTarget; got != PutDefaultMatchTarget {
		t.Errorf("normalized 0 -> %d, want %d", got, PutDefaultMatchTarget)
	}
	if got := (PutConfig{MatchTarget: 999}).normalized().MatchTarget; got != PutDefaultMatchTarget {
		t.Errorf("normalized 999 -> %d, want %d", got, PutDefaultMatchTarget)
	}
	if got := (PutConfig{MatchTarget: 30}).normalized().MatchTarget; got != 30 {
		t.Errorf("normalized 30 -> %d, want 30", got)
	}
}

func TestPutConfigUnmarshalLegacyConfig(t *testing.T) {
	var cfg PutConfig
	if err := json.Unmarshal([]byte(`{"cd":0,"mt":25}`), &cfg); err != nil {
		t.Fatalf("unmarshal legacy config: %v", err)
	}
	if cfg.MatchTarget != 25 {
		t.Errorf("MatchTarget = %d, want 25", cfg.MatchTarget)
	}
}
