//go:build !js || !wasm || classic

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBrusquembilleConfig_Default(t *testing.T) {
	cfg := domain.DefaultBrusquembilleConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() default = %v", err)
	}
}

func TestBrusquembilleConfig_ValidatePlayerCnt(t *testing.T) {
	tests := []struct {
		name    string
		cnt     int
		wantErr bool
	}{
		{"in range", 5, false}, {"too few", 1, true}, {"too many", 6, true}, {"unset", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (domain.BrusquembilleConfig{PlayerCnt: tt.cnt}).Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
