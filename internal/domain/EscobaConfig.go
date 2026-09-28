//go:build !js || !wasm || extra7

package domain

// EscobaDefaultTargetScore 試合終了スコア (先に到達したプレイヤーが勝利)
const EscobaDefaultTargetScore = 10

// EscobaMaxTargetScore Validate で許容する TargetScore の上限
const EscobaMaxTargetScore = 100

// EscobaConfig Escoba ゲーム設定
type EscobaConfig struct {
	TargetScore int `json:"ts"` // 試合終了スコア (デフォルト 10)
}

// DefaultEscobaConfig デフォルト設定を返す
func DefaultEscobaConfig() EscobaConfig {
	return EscobaConfig{
		TargetScore: EscobaDefaultTargetScore,
	}
}

// Validate 設定値のドメインバリデーション
func (c EscobaConfig) Validate() error {
	if err := ValidateRange("target score", c.TargetScore, 1, EscobaMaxTargetScore); err != nil {
		return err
	}
	return nil
}
