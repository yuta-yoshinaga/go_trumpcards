//go:build !js || !wasm || extra4

package domain

// EstimationRoundsMin ラウンド数の下限
const EstimationRoundsMin = 1

// EstimationRoundsMax ラウンド数の上限
const EstimationRoundsMax = 18

// EstimationRoundsDefault 既定のラウンド数（本場の「ファスト」は 5 ラウンド）
const EstimationRoundsDefault = 5

// EstimationConfig エスティメーション ゲーム設定
type EstimationConfig struct {
	// Rounds ゲーム終了までのラウンド数
	Rounds int `json:"rd"`
}

// DefaultEstimationConfig デフォルト設定を返す
func DefaultEstimationConfig() EstimationConfig {
	return EstimationConfig{Rounds: EstimationRoundsDefault}
}

// Validate 設定値のドメインバリデーション
func (c EstimationConfig) Validate() error {
	return ValidateRange("rounds", c.Rounds, EstimationRoundsMin, EstimationRoundsMax)
}
