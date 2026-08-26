//go:build !js || !wasm || extra4

package domain

import "testing"

func TestDefaultPutConfig(t *testing.T) {
	c := DefaultPutConfig()
	if c.CpuDifficulty != PutCpuDifficultyNormal {
		t.Errorf("CpuDifficulty = %d, want Normal", c.CpuDifficulty)
	}
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
		{"bad difficulty", PutConfig{CpuDifficulty: 99, MatchTarget: 15}, true},
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
