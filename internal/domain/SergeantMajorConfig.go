//go:build !js || !wasm || extra

package domain

// SergeantMajorRoundsMin / SergeantMajorRoundsMax は許容するラウンド数の範囲。
//
// **役割が 3 つあるので 3 の倍数が自然。** 3 ラウンドで全員が 8・5・3 を
// 一度ずつ引き受けます。
const (
	SergeantMajorRoundsMin = 3
	SergeantMajorRoundsMax = 30
)

// SergeantMajorConfig はサージェントメジャーのゲーム設定。
type SergeantMajorConfig struct {
	// Rounds は打つラウンド数。
	Rounds int `json:"r"`
}

// DefaultSergeantMajorConfig はデフォルト設定を返す。
func DefaultSergeantMajorConfig() SergeantMajorConfig {
	return SergeantMajorConfig{Rounds: SergeantMajorDefaultRounds}
}

// Validate は設定値のドメインバリデーション。
func (c SergeantMajorConfig) Validate() error {
	return ValidateRange("rounds", c.Rounds, SergeantMajorRoundsMin, SergeantMajorRoundsMax)
}
