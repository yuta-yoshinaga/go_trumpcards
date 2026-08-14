//go:build !js || !wasm || solo

package domain

// TeenDoPaanchRoundsMin / TeenDoPaanchRoundsMax は許容するラウンド数の範囲。
//
// **役割が 3 つあるので 3 の倍数が自然。** 3 ラウンドで全員が 3・2・5 を
// 一度ずつ引き受けます。
const (
	TeenDoPaanchRoundsMin = 3
	TeenDoPaanchRoundsMax = 30
)

// TeenDoPaanchConfig は 3-2-5 のゲーム設定。
type TeenDoPaanchConfig struct {
	// Rounds は打つラウンド数。
	Rounds int `json:"r"`
}

// DefaultTeenDoPaanchConfig はデフォルト設定を返す。
func DefaultTeenDoPaanchConfig() TeenDoPaanchConfig {
	return TeenDoPaanchConfig{Rounds: TeenDoPaanchDefaultRounds}
}

// Validate は設定値のドメインバリデーション。
func (c TeenDoPaanchConfig) Validate() error {
	return ValidateRange("rounds", c.Rounds, TeenDoPaanchRoundsMin, TeenDoPaanchRoundsMax)
}
