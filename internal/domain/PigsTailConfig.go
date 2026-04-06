package domain

import "encoding/json"

// PigsTailConfig ぶたのしっぽ設定
type PigsTailConfig struct {
	CpuHesitationEnabled bool // CPU迷い時間ディレイ
}

// DefaultPigsTailConfig デフォルト設定を返す
func DefaultPigsTailConfig() PigsTailConfig {
	return PigsTailConfig{}
}

// Validate 設定値のドメインバリデーション
func (c PigsTailConfig) Validate() error {
	return nil
}

// pigsTailConfigJSON is the JSON wire format for PigsTailConfig.
type pigsTailConfigJSON struct {
	CpuHesitationEnabled bool `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (c PigsTailConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(pigsTailConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *PigsTailConfig) UnmarshalJSON(data []byte) error {
	var j pigsTailConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuHesitationEnabled = j.CpuHesitationEnabled
	return nil
}
