//go:build !js || !wasm || solo

package domain

import "encoding/json"

// ZhengCpuDifficulty CPU難易度レベル
type ZhengCpuDifficulty int

// ZhengCpuDifficulty定数
const (
	ZhengDifficultyNormal ZhengCpuDifficulty = 0
	ZhengDifficultyEasy   ZhengCpuDifficulty = 1
	ZhengDifficultyHard   ZhengCpuDifficulty = 2
)

// ZhengConfig 争上游ゲーム設定
type ZhengConfig struct {
	CpuDifficulty ZhengCpuDifficulty
}

// DefaultZhengConfig デフォルト設定
func DefaultZhengConfig() ZhengConfig {
	return ZhengConfig{
		CpuDifficulty: ZhengDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c ZhengConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ZhengDifficultyNormal), int(ZhengDifficultyHard))
}

// zhengConfigJSON is the JSON wire format for ZhengConfig.
type zhengConfigJSON struct {
	CpuDifficulty ZhengCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c ZhengConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(zhengConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ZhengConfig) UnmarshalJSON(data []byte) error {
	var j zhengConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	return nil
}
