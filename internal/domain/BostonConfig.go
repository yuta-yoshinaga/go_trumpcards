//go:build !js || !wasm || extra3

package domain

// BostonTargetHandsDefault は既定の規定局数。
const BostonTargetHandsDefault = 8

// BostonTargetHandsMin / BostonTargetHandsMax は規定局数の範囲。
const (
	BostonTargetHandsMin = 1
	BostonTargetHandsMax = 30
)

// BostonConfig ボストンのゲーム設定
type BostonConfig struct {
	// TargetHands は決着までの局数。
	TargetHands int `json:"th"`
}

// DefaultBostonConfig デフォルト設定を返す
func DefaultBostonConfig() BostonConfig {
	return BostonConfig{
		TargetHands: BostonTargetHandsDefault,
	}
}

// Validate 設定値のドメインバリデーション
func (c BostonConfig) Validate() error {
	return ValidateRange("target hands", c.TargetHands, BostonTargetHandsMin, BostonTargetHandsMax)
}
