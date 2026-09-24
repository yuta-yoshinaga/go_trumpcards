//go:build !js || !wasm || extra7

package domain

// ScoponeDefaultTargetScore 試合終了スコア (先に到達したチームが勝利)
const ScoponeDefaultTargetScore = 11

// ScoponeMaxTargetScore Validate で許容する TargetScore の上限
const ScoponeMaxTargetScore = 100

// ScoponeConfig Scopone ゲーム設定
type ScoponeConfig struct {
	TargetScore int `json:"ts"` // 試合終了スコア (デフォルト 11)
}

// DefaultScoponeConfig デフォルト設定を返す
func DefaultScoponeConfig() ScoponeConfig {
	return ScoponeConfig{
		TargetScore: ScoponeDefaultTargetScore,
	}
}

// Validate 設定値のドメインバリデーション
func (c ScoponeConfig) Validate() error {
	if err := ValidateRange("target score", c.TargetScore, 1, ScoponeMaxTargetScore); err != nil {
		return err
	}
	return nil
}
