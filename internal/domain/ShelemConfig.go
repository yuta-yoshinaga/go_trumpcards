//go:build !js || !wasm || extra2

package domain

// ShelemTargetMin 勝利に必要な累計点の下限
const ShelemTargetMin = 100

// ShelemTargetMax 勝利に必要な累計点の上限
const ShelemTargetMax = 2000

// ShelemTargetDefault 既定の勝利点
const ShelemTargetDefault = 500

// ShelemConfig シェレム ゲーム設定
type ShelemConfig struct {
	// Target 勝利に必要な累計点
	Target int `json:"tg"`
}

// DefaultShelemConfig デフォルト設定を返す
func DefaultShelemConfig() ShelemConfig {
	return ShelemConfig{Target: ShelemTargetDefault}
}

// Validate 設定値のドメインバリデーション
func (c ShelemConfig) Validate() error {
	return ValidateRange("target", c.Target, ShelemTargetMin, ShelemTargetMax)
}
