//go:build !js || !wasm || extra3

package domain

// TarabishTargetMin 勝利点の下限
const TarabishTargetMin = 100

// TarabishTargetMax 勝利点の上限
const TarabishTargetMax = 1000

// TarabishTargetDefault 既定の勝利点（本場のタラビッシュは 500 点先取）
const TarabishTargetDefault = 500

// TarabishConfig タラビッシュ ゲーム設定
type TarabishConfig struct {
	// Target 勝利に必要な点数
	Target int `json:"tg"`
}

// DefaultTarabishConfig デフォルト設定を返す
func DefaultTarabishConfig() TarabishConfig {
	return TarabishConfig{Target: TarabishTargetDefault}
}

// Validate 設定値のドメインバリデーション
func (c TarabishConfig) Validate() error {
	return ValidateRange("target", c.Target, TarabishTargetMin, TarabishTargetMax)
}
