//go:build !js || !wasm || extra2

package domain

// BalootTargetMin 勝利点の下限
const BalootTargetMin = 50

// BalootTargetMax 勝利点の上限
const BalootTargetMax = 500

// BalootTargetDefault 既定の勝利点（本場のバルートは 152 点先取）
const BalootTargetDefault = 152

// BalootConfig バルート ゲーム設定
type BalootConfig struct {
	// Target 勝利に必要な点数
	Target int `json:"tg"`
}

// DefaultBalootConfig デフォルト設定を返す
func DefaultBalootConfig() BalootConfig {
	return BalootConfig{Target: BalootTargetDefault}
}

// Validate 設定値のドメインバリデーション
func (c BalootConfig) Validate() error {
	return ValidateRange("target", c.Target, BalootTargetMin, BalootTargetMax)
}
