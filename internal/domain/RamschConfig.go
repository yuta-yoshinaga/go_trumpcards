//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// RamschCpuDifficulty Ramsch CPU difficulty level
type RamschCpuDifficulty int

// Ramsch CPU difficulty constants
const (
	// RamschCpuDifficultyEasy easy
	RamschCpuDifficultyEasy RamschCpuDifficulty = iota
	// RamschCpuDifficultyNormal normal
	RamschCpuDifficultyNormal
	// RamschCpuDifficultyHard hard
	RamschCpuDifficultyHard
)

// RamschConfig Ramsch game configuration
type RamschConfig struct {
	CpuDifficulty RamschCpuDifficulty
	// TargetScore は**罰点の上限**。累計失点がこれに達した人が出た時点で対局終了で、
	// 勝つのは残りのうち失点が最も少ない人。到達した人が勝つのではない。
	TargetScore int
}

// DefaultRamschConfig returns the default configuration
func DefaultRamschConfig() RamschConfig {
	return RamschConfig{
		CpuDifficulty: RamschCpuDifficultyNormal,
		TargetScore:   500,
	}
}

// Validate domain validation for the configuration
func (c RamschConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(RamschCpuDifficultyEasy), int(RamschCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target score", c.TargetScore, 1); err != nil {
		return err
	}
	return nil
}

// ramschConfigJSON is the JSON wire format for RamschConfig.
type ramschConfigJSON struct {
	CpuDifficulty RamschCpuDifficulty `json:"cd"`
	TargetScore   int                 `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (c RamschConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(ramschConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *RamschConfig) UnmarshalJSON(data []byte) error {
	var j ramschConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	c.TargetScore = j.TargetScore
	return nil
}
