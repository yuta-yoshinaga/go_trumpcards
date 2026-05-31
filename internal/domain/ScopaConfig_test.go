package domain

import (
	"encoding/json"
	"testing"
)

func TestDefaultScopaConfig(t *testing.T) {
	c := DefaultScopaConfig()
	if c.TargetScore != ScopaDefaultTargetScore {
		t.Errorf("target = %d, want %d", c.TargetScore, ScopaDefaultTargetScore)
	}
	if c.CpuDifficulty != ScopaDifficultyNormal {
		t.Errorf("difficulty = %d, want Normal", c.CpuDifficulty)
	}
	if err := c.Validate(); err != nil {
		t.Errorf("default config should validate: %v", err)
	}
}

func TestScopaConfigValidate(t *testing.T) {
	cases := []struct {
		name string
		cfg  ScopaConfig
		ok   bool
	}{
		{"valid", ScopaConfig{TargetScore: 11, CpuDifficulty: ScopaDifficultyHard}, true},
		{"difficulty too high", ScopaConfig{TargetScore: 11, CpuDifficulty: 5}, false},
		{"difficulty negative", ScopaConfig{TargetScore: 11, CpuDifficulty: -1}, false},
		{"target too low", ScopaConfig{TargetScore: 0, CpuDifficulty: ScopaDifficultyNormal}, false},
		{"target too high", ScopaConfig{TargetScore: 1000, CpuDifficulty: ScopaDifficultyNormal}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.cfg.Validate()
			if c.ok && err != nil {
				t.Errorf("expected valid, got %v", err)
			}
			if !c.ok && err == nil {
				t.Error("expected invalid")
			}
		})
	}
}

func TestScopaConfigJSONRoundTrip(t *testing.T) {
	c := ScopaConfig{TargetScore: 21, CpuDifficulty: ScopaDifficultyHard}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got ScopaConfig
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != c {
		t.Errorf("round-trip mismatch: %+v vs %+v", got, c)
	}
}

func TestScopaDifficultyNames(t *testing.T) {
	for _, d := range []ScopaCpuDifficulty{ScopaDifficultyEasy, ScopaDifficultyNormal, ScopaDifficultyHard} {
		if ScopaDifficultyNames[d] == "" {
			t.Errorf("missing name for difficulty %d", d)
		}
	}
}
