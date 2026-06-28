package domain

import "encoding/json"

// Beggar-My-Neighbour ゲーム設定の範囲定数
const (
	// BeggarMyNeighbourMinMaxRounds MaxRounds の下限
	BeggarMyNeighbourMinMaxRounds = 10
	// BeggarMyNeighbourMaxMaxRounds MaxRounds の上限
	BeggarMyNeighbourMaxMaxRounds = 50000
	// BeggarMyNeighbourDefaultMaxRounds MaxRounds のデフォルト値
	BeggarMyNeighbourDefaultMaxRounds = 2000
)

// BeggarMyNeighbourConfig Beggar-My-Neighbour ゲーム設定
type BeggarMyNeighbourConfig struct {
	// MaxRounds ラウンド上限 (到達時は保有枚数の多い方を勝者とする)
	MaxRounds int
}

// DefaultBeggarMyNeighbourConfig デフォルト設定を返す
func DefaultBeggarMyNeighbourConfig() BeggarMyNeighbourConfig {
	return BeggarMyNeighbourConfig{MaxRounds: BeggarMyNeighbourDefaultMaxRounds}
}

// Validate 設定値の検証
func (c BeggarMyNeighbourConfig) Validate() error {
	return ValidateRange("max rounds", c.MaxRounds, BeggarMyNeighbourMinMaxRounds, BeggarMyNeighbourMaxMaxRounds)
}

// beggarMyNeighbourConfigJSON is the JSON wire format for BeggarMyNeighbourConfig.
type beggarMyNeighbourConfigJSON struct {
	MaxRounds int `json:"mr"`
}

// MarshalJSON implements json.Marshaler.
func (c BeggarMyNeighbourConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(beggarMyNeighbourConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *BeggarMyNeighbourConfig) UnmarshalJSON(data []byte) error {
	var j beggarMyNeighbourConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.MaxRounds = j.MaxRounds
	return nil
}
