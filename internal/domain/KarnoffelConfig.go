//go:build !js || !wasm || classic

package domain

// 勝利に要る局数の範囲。
const (
	// KarnoffelMinTarget 最少
	KarnoffelMinTarget = 1
	// KarnoffelMaxTarget 最多
	KarnoffelMaxTarget = 10
)

// KarnoffelConfig カルニッフェルのゲーム設定
type KarnoffelConfig struct {
	// TargetHands は勝利に要る局数。
	TargetHands int `json:"th"`
}

// DefaultKarnoffelConfig デフォルト設定を返す
func DefaultKarnoffelConfig() KarnoffelConfig {
	return KarnoffelConfig{
		TargetHands: KarnoffelDefaultTarget,
	}
}

// Validate 設定値のドメインバリデーション
func (c KarnoffelConfig) Validate() error {
	return ValidateRange("target hands", c.TargetHands, KarnoffelMinTarget, KarnoffelMaxTarget)
}
