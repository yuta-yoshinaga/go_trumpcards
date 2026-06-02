package domain

import "encoding/json"

// TienLenCpuDifficulty CPU難易度レベル
type TienLenCpuDifficulty int

// TienLenCpuDifficulty定数
const (
	TienLenDifficultyNormal TienLenCpuDifficulty = 0
	TienLenDifficultyEasy   TienLenCpuDifficulty = 1
	TienLenDifficultyHard   TienLenCpuDifficulty = 2
)

// TienLenConfig Tien Lenゲーム設定
type TienLenConfig struct {
	CpuDifficulty TienLenCpuDifficulty
}

// DefaultTienLenConfig デフォルト設定
func DefaultTienLenConfig() TienLenConfig {
	return TienLenConfig{
		CpuDifficulty: TienLenDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c TienLenConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TienLenDifficultyNormal), int(TienLenDifficultyHard))
}

// tienLenConfigJSON is the JSON wire format for TienLenConfig.
type tienLenConfigJSON struct {
	CpuDifficulty TienLenCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c TienLenConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(tienLenConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *TienLenConfig) UnmarshalJSON(data []byte) error {
	var j tienLenConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	return nil
}
