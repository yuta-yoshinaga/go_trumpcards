//go:build !js || !wasm || extra4

package domain

// 1 ゲームの局数の範囲。
const (
	// SixBidSoloMinHands 最少局数
	SixBidSoloMinHands = 3
	// SixBidSoloMaxHands 最多局数
	SixBidSoloMaxHands = 12
	// SixBidSoloDefaultHands 既定の局数
	SixBidSoloDefaultHands = 6
)

// SixBidSoloConfig シックスビッド・ソロのゲーム設定
type SixBidSoloConfig struct {
	// TargetHands は 1 ゲームの局数。
	//
	// **勝利は点数の先取ではなく規定局数後の首位。**受け払いが差分式で
	// 額が大きく振れるので、点数目標にすると 1 局で決着してしまう。
	TargetHands int `json:"th"`
}

// DefaultSixBidSoloConfig デフォルト設定を返す
func DefaultSixBidSoloConfig() SixBidSoloConfig {
	return SixBidSoloConfig{
		TargetHands: SixBidSoloDefaultHands,
	}
}

// Validate 設定値のドメインバリデーション
func (c SixBidSoloConfig) Validate() error {
	return ValidateRange("target hands", c.TargetHands, SixBidSoloMinHands, SixBidSoloMaxHands)
}
