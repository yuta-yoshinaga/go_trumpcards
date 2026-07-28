//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// FaroConfig はファロの設定を表す。
type FaroConfig struct {
	// StartChips は開始時のチップ数。
	StartChips int
	// MinBet は1ランクあたりの最低ベット額。
	MinBet int
	// MaxBet は1ランクあたりの最大ベット額。
	MaxBet int
}

// Faro の設定既定値とドメイン上限。
const (
	// FaroDefaultStartChips はデフォルトの開始チップ数。
	FaroDefaultStartChips = 1000
	// FaroDefaultMinBet はデフォルトの最低ベット額。
	FaroDefaultMinBet = 10
	// FaroDefaultMaxBet はデフォルトの最大ベット額。
	FaroDefaultMaxBet = 10000
	// FaroChipsUpperBound は StartChips の妥当性検証に使う上限。
	FaroChipsUpperBound = 1000000000
)

// DefaultFaroConfig はデフォルト設定を返す。
func DefaultFaroConfig() FaroConfig {
	return FaroConfig{
		StartChips: FaroDefaultStartChips,
		MinBet:     FaroDefaultMinBet,
		MaxBet:     FaroDefaultMaxBet,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c FaroConfig) Validate() error {
	if err := ValidateRange("start chips", c.StartChips, 1, FaroChipsUpperBound); err != nil {
		return err
	}
	if err := ValidateRange("min bet", c.MinBet, 1, FaroChipsUpperBound); err != nil {
		return err
	}
	if err := ValidateRange("max bet", c.MaxBet, c.MinBet, FaroChipsUpperBound); err != nil {
		return err
	}
	return nil
}

// faroConfigJSON は FaroConfig の JSON ワイヤーフォーマット。
type faroConfigJSON struct {
	StartChips int `json:"sc"`
	MinBet     int `json:"mn"`
	MaxBet     int `json:"mx"`
}

// MarshalJSON implements json.Marshaler.
func (c FaroConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(faroConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *FaroConfig) UnmarshalJSON(data []byte) error {
	var j faroConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.StartChips = j.StartChips
	c.MinBet = j.MinBet
	c.MaxBet = j.MaxBet
	return nil
}
