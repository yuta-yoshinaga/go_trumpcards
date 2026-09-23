//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func TestDefaultPiquetConfig(t *testing.T) {
	c := DefaultPiquetConfig()
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
		{"deals zero", PiquetConfig{DealsPerPartie: 0}, true},
		{"deals negative", PiquetConfig{DealsPerPartie: -1}, true},
		{"single deal ok", PiquetConfig{DealsPerPartie: 1}, false},
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
	orig := PiquetConfig{DealsPerPartie: 3}
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

func TestPiquetConfigUnmarshalLegacyDifficulty(t *testing.T) {
	var got PiquetConfig
	if err := json.Unmarshal([]byte(`{"cd":0,"dp":3}`), &got); err != nil {
		t.Fatalf("Unmarshal legacy config: %v", err)
	}
	if got.DealsPerPartie != 3 {
		t.Errorf("DealsPerPartie = %d, want 3", got.DealsPerPartie)
	}
}
