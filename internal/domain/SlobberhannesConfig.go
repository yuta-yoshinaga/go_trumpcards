//go:build !js || !wasm || classic

package domain

// SlobberhannesRoundsMin ラウンド数の下限
const SlobberhannesRoundsMin = 1

// SlobberhannesRoundsMax ラウンド数の上限
const SlobberhannesRoundsMax = 12

// SlobberhannesRoundsDefault 既定のラウンド数。4 人が 1 回ずつリードを
// 引き受けて一巡する回数。
const SlobberhannesRoundsDefault = 4

// SlobberhannesConfig スロバーハンネス ゲーム設定
type SlobberhannesConfig struct {
	// Rounds 何ラウンドで打ち切るか
	Rounds int `json:"rd"`
}

// DefaultSlobberhannesConfig デフォルト設定を返す
func DefaultSlobberhannesConfig() SlobberhannesConfig {
	return SlobberhannesConfig{Rounds: SlobberhannesRoundsDefault}
}

// Validate 設定値のドメインバリデーション
func (c SlobberhannesConfig) Validate() error {
	return ValidateRange("rounds", c.Rounds, SlobberhannesRoundsMin, SlobberhannesRoundsMax)
}
