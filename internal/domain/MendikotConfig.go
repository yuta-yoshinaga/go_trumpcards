//go:build !js || !wasm || extra4

package domain

// MendikotTargetMin 勝利に必要なハンド勝ち点の下限
const MendikotTargetMin = 1

// MendikotTargetMax 勝利に必要なハンド勝ち点の上限
const MendikotTargetMax = 20

// MendikotTargetDefault 既定の勝利ハンド勝ち点
const MendikotTargetDefault = 5

// MendikotConfig メンディコット ゲーム設定
type MendikotConfig struct {
	// Target 勝利に必要なハンド勝ち点
	Target int `json:"tg"`
}

// DefaultMendikotConfig デフォルト設定を返す
func DefaultMendikotConfig() MendikotConfig {
	return MendikotConfig{Target: MendikotTargetDefault}
}

// Validate 設定値のドメインバリデーション
func (c MendikotConfig) Validate() error {
	return ValidateRange("target", c.Target, MendikotTargetMin, MendikotTargetMax)
}
