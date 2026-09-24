//go:build !js || !wasm || extra6

package domain

// EcarteDefaultTargetScore 試合終了スコア (先に到達した側が勝利)
const EcarteDefaultTargetScore = 5

// EcarteMaxTargetScore Validate で許容する TargetScore の上限
const EcarteMaxTargetScore = 50

// EcarteConfig Écarté ゲーム設定
type EcarteConfig struct {
	TargetScore int `json:"ts"` // 試合終了スコア (デフォルト 5)
}

// DefaultEcarteConfig デフォルト設定を返す
func DefaultEcarteConfig() EcarteConfig {
	return EcarteConfig{
		TargetScore: EcarteDefaultTargetScore,
	}
}

// Validate 設定値のドメインバリデーション
func (c EcarteConfig) Validate() error {
	if err := ValidateRange("target score", c.TargetScore, 1, EcarteMaxTargetScore); err != nil {
		return err
	}
	return nil
}
