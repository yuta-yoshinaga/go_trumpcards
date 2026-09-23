//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// BassetConfig はバセットの設定を表す。
type BassetConfig struct {
	// StartChips は開始時のチップ数。
	StartChips int
	// MinBet は1ランクあたりの最低ベット額。
	MinBet int
	// MaxBet は1ランクあたりの最大ベット額。
	MaxBet int
}

// Basset の設定既定値とドメイン上限。
const (
	// BassetDefaultStartChips はデフォルトの開始チップ数。
	BassetDefaultStartChips = 1000
	// BassetDefaultMinBet はデフォルトの最低ベット額。
	BassetDefaultMinBet = 10
	// BassetDefaultMaxBet はデフォルトの最大ベット額。
	BassetDefaultMaxBet = 10000
	// BassetChipsUpperBound は StartChips の妥当性検証に使う上限。
	BassetChipsUpperBound = 1000000000
)

// DefaultBassetConfig はデフォルト設定を返す。
func DefaultBassetConfig() BassetConfig {
	return BassetConfig{
		StartChips: BassetDefaultStartChips,
		MinBet:     BassetDefaultMinBet,
		MaxBet:     BassetDefaultMaxBet,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c BassetConfig) Validate() error {
	if err := ValidateRange("start chips", c.StartChips, 1, BassetChipsUpperBound); err != nil {
		return err
	}
	if err := ValidateRange("min bet", c.MinBet, 1, BassetChipsUpperBound); err != nil {
		return err
	}
	if err := ValidateRange("max bet", c.MaxBet, c.MinBet, BassetChipsUpperBound); err != nil {
		return err
	}
	return nil
}

// bassetConfigJSON は BassetConfig の JSON ワイヤーフォーマット。
type bassetConfigJSON struct {
	StartChips int `json:"sc"`
	MinBet     int `json:"mn"`
	MaxBet     int `json:"mx"`
}

// MarshalJSON implements json.Marshaler.
func (c BassetConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(bassetConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *BassetConfig) UnmarshalJSON(data []byte) error {
	var j bassetConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.StartChips = j.StartChips
	c.MinBet = j.MinBet
	c.MaxBet = j.MaxBet
	return nil
}
