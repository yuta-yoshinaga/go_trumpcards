//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// PiquetCpuDifficulty Piquet CPU difficulty level
type PiquetCpuDifficulty int

// Piquet CPU difficulty constants
const (
	// PiquetCpuDifficultyEasy easy
	PiquetCpuDifficultyEasy PiquetCpuDifficulty = iota
	// PiquetCpuDifficultyNormal normal
	PiquetCpuDifficultyNormal
	// PiquetCpuDifficultyHard hard
	PiquetCpuDifficultyHard
)

// PiquetDealsPerPartie 1パルティ (試合) のディール数 (古典的な6ディール)
const PiquetDealsPerPartie = 6

// PiquetConfig Piquet ゲーム設定
type PiquetConfig struct {
	CpuDifficulty  PiquetCpuDifficulty
	DealsPerPartie int // 1パルティのディール数 (default = 6)
}

// DefaultPiquetConfig returns the default configuration
func DefaultPiquetConfig() PiquetConfig {
	return PiquetConfig{
		CpuDifficulty:  PiquetCpuDifficultyNormal,
		DealsPerPartie: PiquetDealsPerPartie,
	}
}

// Validate domain validation for the configuration
func (c PiquetConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(PiquetCpuDifficultyEasy), int(PiquetCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("deals per partie", c.DealsPerPartie, 1); err != nil {
		return err
	}
	return nil
}

// piquetConfigJSON is the JSON wire format for PiquetConfig.
type piquetConfigJSON struct {
	CpuDifficulty  PiquetCpuDifficulty `json:"cd"`
	DealsPerPartie int                 `json:"dp"`
}

// MarshalJSON implements json.Marshaler.
func (c PiquetConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(piquetConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *PiquetConfig) UnmarshalJSON(data []byte) error {
	var j piquetConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	c.DealsPerPartie = j.DealsPerPartie
	return nil
}
