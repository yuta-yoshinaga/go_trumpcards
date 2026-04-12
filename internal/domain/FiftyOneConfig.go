package domain

import "encoding/json"

// FiftyOneCpuDifficulty CPU難易度レベル
type FiftyOneCpuDifficulty int

// FiftyOneCpuDifficulty定数
const (
	// FiftyOneDifficultyEasy 簡単
	FiftyOneDifficultyEasy FiftyOneCpuDifficulty = 0
	// FiftyOneDifficultyNormal 通常 (デフォルト)
	FiftyOneDifficultyNormal FiftyOneCpuDifficulty = 1
	// FiftyOneDifficultyHard 難しい
	FiftyOneDifficultyHard FiftyOneCpuDifficulty = 2
)

// FiftyOneConfig フィフティワン設定
type FiftyOneConfig struct {
	CpuDifficulty FiftyOneCpuDifficulty `json:"-"` // CPU難易度
}

// DefaultFiftyOneConfig デフォルト設定を返す
func DefaultFiftyOneConfig() FiftyOneConfig {
	return FiftyOneConfig{
		CpuDifficulty: FiftyOneDifficultyNormal,
	}
}

// Validate 設定値のバリデーション
func (c FiftyOneConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(FiftyOneDifficultyEasy), int(FiftyOneDifficultyHard))
}

// fiftyOneConfigJSON is the JSON wire format for FiftyOneConfig.
type fiftyOneConfigJSON struct {
	CpuDifficulty FiftyOneCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c FiftyOneConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(fiftyOneConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *FiftyOneConfig) UnmarshalJSON(data []byte) error {
	var j fiftyOneConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	return nil
}
