//go:build !js || !wasm || extra4

package domain

// HoneymoonBridgeTargetMin / HoneymoonBridgeTargetMax は許容する目標点の範囲。
const (
	HoneymoonBridgeTargetMin = 50
	HoneymoonBridgeTargetMax = 500
)

// HoneymoonBridgeConfig はハネムーンブリッジのゲーム設定。
type HoneymoonBridgeConfig struct {
	// Target は勝利に必要な点数。
	Target int `json:"t"`
}

// DefaultHoneymoonBridgeConfig はデフォルト設定を返す。
func DefaultHoneymoonBridgeConfig() HoneymoonBridgeConfig {
	return HoneymoonBridgeConfig{Target: HoneymoonBridgeDefaultTarget}
}

// Validate は設定値のドメインバリデーション。
func (c HoneymoonBridgeConfig) Validate() error {
	return ValidateRange("target", c.Target, HoneymoonBridgeTargetMin, HoneymoonBridgeTargetMax)
}
