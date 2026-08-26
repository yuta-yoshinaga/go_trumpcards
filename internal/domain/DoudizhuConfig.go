//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// DoudizhuCpuDifficulty CPU難易度レベル
type DoudizhuCpuDifficulty int

// DoudizhuCpuDifficulty定数
const (
	// DoudizhuDifficultyNormal 通常難易度 (デフォルト)
	DoudizhuDifficultyNormal DoudizhuCpuDifficulty = 0
	// DoudizhuDifficultyEasy 簡単
	DoudizhuDifficultyEasy DoudizhuCpuDifficulty = 1
	// DoudizhuDifficultyHard 難しい
	DoudizhuDifficultyHard DoudizhuCpuDifficulty = 2
)

// DoudizhuDifficultyNames 難易度名マップ
var DoudizhuDifficultyNames = map[DoudizhuCpuDifficulty]string{
	DoudizhuDifficultyNormal: "Normal",
	DoudizhuDifficultyEasy:   "Easy",
	DoudizhuDifficultyHard:   "Hard",
}

// DoudizhuConfig 斗地主設定
type DoudizhuConfig struct {
	CpuDifficulty DoudizhuCpuDifficulty
}

// DefaultDoudizhuConfig デフォルト設定
func DefaultDoudizhuConfig() DoudizhuConfig {
	return DoudizhuConfig{
		CpuDifficulty: DoudizhuDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c DoudizhuConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(DoudizhuDifficultyNormal), int(DoudizhuDifficultyHard))
}

// doudizhuConfigJSON is the JSON wire format for DoudizhuConfig.
type doudizhuConfigJSON struct {
	CpuDifficulty DoudizhuCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c DoudizhuConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(doudizhuConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *DoudizhuConfig) UnmarshalJSON(data []byte) error {
	var j doudizhuConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	return nil
}
