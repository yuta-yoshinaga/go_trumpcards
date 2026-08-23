//go:build !js || !wasm || classic

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBrusquembilleConfig_Default(t *testing.T) {
	cfg := domain.DefaultBrusquembilleConfig()
	if cfg.CpuDifficulty != domain.BrusquembilleCpuDifficultyNormal {
		t.Errorf("CpuDifficulty = %d, want %d", cfg.CpuDifficulty, domain.BrusquembilleCpuDifficultyNormal)
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() default = %v, want nil", err)
	}
}

func TestBrusquembilleConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.BrusquembilleConfig
		wantErr bool
	}{
		{"default", domain.DefaultBrusquembilleConfig(), false},
		{"normal explicit", domain.BrusquembilleConfig{CpuDifficulty: domain.BrusquembilleCpuDifficultyNormal}, false},
		{"too low", domain.BrusquembilleConfig{CpuDifficulty: -1}, true},
		{"too high", domain.BrusquembilleConfig{CpuDifficulty: 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
