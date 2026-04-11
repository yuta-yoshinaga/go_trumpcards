package domain

import "encoding/json"

// War ゲーム設定の範囲定数
const (
	// WarMinMaxRounds MaxRounds の下限
	WarMinMaxRounds = 10
	// WarMaxMaxRounds MaxRounds の上限
	WarMaxMaxRounds = 10000
	// WarDefaultMaxRounds MaxRounds のデフォルト値
	WarDefaultMaxRounds = 500
)

// WarConfig 戦争ゲーム設定
type WarConfig struct {
	// MaxRounds ラウンド上限 (到達時は保有枚数の多い方を勝者とする)
	MaxRounds int
}

// DefaultWarConfig デフォルト設定を返す
func DefaultWarConfig() WarConfig {
	return WarConfig{MaxRounds: WarDefaultMaxRounds}
}

// Validate 設定値の検証
func (c WarConfig) Validate() error {
	return ValidateRange("max rounds", c.MaxRounds, WarMinMaxRounds, WarMaxMaxRounds)
}

// warConfigJSON is the JSON wire format for WarConfig.
type warConfigJSON struct {
	MaxRounds int `json:"mr"`
}

// MarshalJSON implements json.Marshaler.
func (c WarConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(warConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *WarConfig) UnmarshalJSON(data []byte) error {
	var j warConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.MaxRounds = j.MaxRounds
	return nil
}
