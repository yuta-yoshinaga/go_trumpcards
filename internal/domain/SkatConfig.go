//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// SkatCpuDifficulty Skat CPU difficulty level
type SkatCpuDifficulty int

// Skat CPU difficulty constants
const (
	// SkatCpuDifficultyEasy easy
	SkatCpuDifficultyEasy SkatCpuDifficulty = iota
	// SkatCpuDifficultyNormal normal
	SkatCpuDifficultyNormal
	// SkatCpuDifficultyHard hard
	SkatCpuDifficultyHard
)

// SkatConfig Skat game configuration
type SkatConfig struct {
	CpuDifficulty SkatCpuDifficulty
	TargetScore   int // game-end target score (first to reach wins)
}

// DefaultSkatConfig returns the default configuration
func DefaultSkatConfig() SkatConfig {
	return SkatConfig{
		CpuDifficulty: SkatCpuDifficultyNormal,
		TargetScore:   500,
	}
}

// Validate domain validation for the configuration
func (c SkatConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SkatCpuDifficultyEasy), int(SkatCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target score", c.TargetScore, 1); err != nil {
		return err
	}
	return nil
}

// skatConfigJSON is the JSON wire format for SkatConfig.
type skatConfigJSON struct {
	CpuDifficulty SkatCpuDifficulty `json:"cd"`
	TargetScore   int               `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (c SkatConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(skatConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *SkatConfig) UnmarshalJSON(data []byte) error {
	var j skatConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	c.TargetScore = j.TargetScore
	return nil
}
