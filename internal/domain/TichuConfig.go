//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// TichuCpuDifficulty CPU難易度レベル
type TichuCpuDifficulty int

// TichuCpuDifficulty定数
const (
	// TichuDifficultyNormal 通常難易度 (デフォルト)
	TichuDifficultyNormal TichuCpuDifficulty = 0
	// TichuDifficultyEasy 簡単
	TichuDifficultyEasy TichuCpuDifficulty = 1
	// TichuDifficultyHard 難しい
	TichuDifficultyHard TichuCpuDifficulty = 2
)

// TichuDifficultyNames 難易度名マップ
var TichuDifficultyNames = map[TichuCpuDifficulty]string{
	TichuDifficultyNormal: "Normal",
	TichuDifficultyEasy:   "Easy",
	TichuDifficultyHard:   "Hard",
}

// TichuConfig ティチュー設定
type TichuConfig struct {
	CpuDifficulty TichuCpuDifficulty
}

// DefaultTichuConfig デフォルト設定
func DefaultTichuConfig() TichuConfig {
	return TichuConfig{
		CpuDifficulty: TichuDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c TichuConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TichuDifficultyNormal), int(TichuDifficultyHard))
}

// tichuConfigJSON is the JSON wire format for TichuConfig.
type tichuConfigJSON struct {
	CpuDifficulty TichuCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c TichuConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(tichuConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *TichuConfig) UnmarshalJSON(data []byte) error {
	var j tichuConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	return nil
}
