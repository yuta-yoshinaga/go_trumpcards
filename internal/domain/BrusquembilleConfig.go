//go:build !js || !wasm || extra7

package domain

// BrusquembilleConfig ブリュスカンビーユゲーム設定
type BrusquembilleConfig struct {
	// PlayerCnt 席数 (2-5)。席 0 が人間。
	PlayerCnt int `json:"pc"`
}

// DefaultBrusquembilleConfig デフォルト設定を返す
func DefaultBrusquembilleConfig() BrusquembilleConfig {
	return BrusquembilleConfig{
		PlayerCnt: BrusquembilleDefaultPlayerCnt,
	}
}

// Validate 設定値のドメインバリデーション
func (c BrusquembilleConfig) Validate() error {
	// **席数も検査する。** ここを素通しにすると、範囲外の席数がそのまま
	// 設定に入り、卓が組めないまま Reset される。
	return ValidateRange("player count", c.PlayerCnt,
		BrusquembilleMinPlayerCnt, BrusquembilleMaxPlayerCnt)
}
