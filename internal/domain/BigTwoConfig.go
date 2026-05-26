package domain

import "encoding/json"

// BigTwoCpuDifficulty CPU難易度レベル
type BigTwoCpuDifficulty int

// BigTwoCpuDifficulty定数
const (
	BigTwoDifficultyNormal BigTwoCpuDifficulty = 0
	BigTwoDifficultyEasy   BigTwoCpuDifficulty = 1
	BigTwoDifficultyHard   BigTwoCpuDifficulty = 2
)

// BigTwoConfig Big Twoゲーム設定
type BigTwoConfig struct {
	CpuDifficulty BigTwoCpuDifficulty
}

// DefaultBigTwoConfig デフォルト設定
func DefaultBigTwoConfig() BigTwoConfig {
	return BigTwoConfig{
		CpuDifficulty: BigTwoDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c BigTwoConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(BigTwoDifficultyNormal), int(BigTwoDifficultyHard))
}

// bigTwoConfigJSON is the JSON wire format for BigTwoConfig.
type bigTwoConfigJSON struct {
	CpuDifficulty BigTwoCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c BigTwoConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(bigTwoConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *BigTwoConfig) UnmarshalJSON(data []byte) error {
	var j bigTwoConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	return nil
}
