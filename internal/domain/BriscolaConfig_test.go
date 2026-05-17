package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBriscolaConfig_Default(t *testing.T) {
	cfg := domain.DefaultBriscolaConfig()
	if cfg.CpuDifficulty != domain.BriscolaCpuDifficultyNormal {
		t.Errorf("CpuDifficulty = %d, want %d", cfg.CpuDifficulty, domain.BriscolaCpuDifficultyNormal)
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() default = %v, want nil", err)
	}
}

func TestBriscolaConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.BriscolaConfig
		wantErr bool
	}{
		{"default", domain.DefaultBriscolaConfig(), false},
		{"normal explicit", domain.BriscolaConfig{CpuDifficulty: domain.BriscolaCpuDifficultyNormal}, false},
		{"too low", domain.BriscolaConfig{CpuDifficulty: -1}, true},
		{"too high", domain.BriscolaConfig{CpuDifficulty: 1}, true},
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
