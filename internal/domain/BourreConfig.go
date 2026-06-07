//go:build !js || !wasm || casino

package domain

import "encoding/json"

// BourreCpuDifficulty CPU難易度レベル
type BourreCpuDifficulty int

// BourreCpuDifficulty定数
const (
	// BourreDifficultyNormal 通常難易度 (デフォルト)
	BourreDifficultyNormal BourreCpuDifficulty = 0
	// BourreDifficultyEasy 簡単
	BourreDifficultyEasy BourreCpuDifficulty = 1
	// BourreDifficultyHard 難しい
	BourreDifficultyHard BourreCpuDifficulty = 2
)

// BourreDifficultyNames 難易度名マップ
var BourreDifficultyNames = map[BourreCpuDifficulty]string{
	BourreDifficultyNormal: "Normal",
	BourreDifficultyEasy:   "Easy",
	BourreDifficultyHard:   "Hard",
}

// BourreConfig ブーレ設定
type BourreConfig struct {
	CpuDifficulty BourreCpuDifficulty
}

// DefaultBourreConfig デフォルト設定
func DefaultBourreConfig() BourreConfig {
	return BourreConfig{
		CpuDifficulty: BourreDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c BourreConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(BourreDifficultyNormal), int(BourreDifficultyHard))
}

// bourreConfigJSON is the JSON wire format for BourreConfig.
type bourreConfigJSON struct {
	CpuDifficulty BourreCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c BourreConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(bourreConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *BourreConfig) UnmarshalJSON(data []byte) error {
	var j bourreConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	return nil
}
