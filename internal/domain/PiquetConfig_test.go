//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func TestDefaultPiquetConfig(t *testing.T) {
	c := DefaultPiquetConfig()
	if c.CpuDifficulty != PiquetCpuDifficultyNormal {
		t.Errorf("default CpuDifficulty = %v, want Normal", c.CpuDifficulty)
	}
	if c.DealsPerPartie != 6 {
		t.Errorf("default DealsPerPartie = %d, want 6", c.DealsPerPartie)
	}
	if err := c.Validate(); err != nil {
		t.Errorf("default config Validate() = %v, want nil", err)
	}
}

func TestPiquetConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  PiquetConfig
		wantErr bool
	}{
		{"default ok", DefaultPiquetConfig(), false},
		{"easy ok", PiquetConfig{CpuDifficulty: PiquetCpuDifficultyEasy, DealsPerPartie: 6}, false},
		{"hard ok", PiquetConfig{CpuDifficulty: PiquetCpuDifficultyHard, DealsPerPartie: 6}, false},
		{"difficulty below range", PiquetConfig{CpuDifficulty: PiquetCpuDifficulty(-1), DealsPerPartie: 6}, true},
		{"difficulty above range", PiquetConfig{CpuDifficulty: PiquetCpuDifficulty(99), DealsPerPartie: 6}, true},
		{"deals zero", PiquetConfig{CpuDifficulty: PiquetCpuDifficultyNormal, DealsPerPartie: 0}, true},
		{"deals negative", PiquetConfig{CpuDifficulty: PiquetCpuDifficultyNormal, DealsPerPartie: -1}, true},
		{"single deal ok", PiquetConfig{CpuDifficulty: PiquetCpuDifficultyNormal, DealsPerPartie: 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPiquetConfigJSONRoundTrip(t *testing.T) {
	orig := PiquetConfig{CpuDifficulty: PiquetCpuDifficultyHard, DealsPerPartie: 3}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got PiquetConfig
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != orig {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, orig)
	}
}
