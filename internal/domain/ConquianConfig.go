//go:build !js || !wasm || extra7

package domain

// ConquianConfig コンキャンゲーム設定
type ConquianConfig struct {
	TargetWins int `json:"tw"` // マッチ勝利数 (先にこの数のラウンドを取ったプレイヤーが勝利)
}

// DefaultConquianConfig デフォルト設定を返す
func DefaultConquianConfig() ConquianConfig {
	return ConquianConfig{
		TargetWins: 1,
	}
}

// Validate 設定値のドメインバリデーション
func (c ConquianConfig) Validate() error {
	if err := ValidateRange("target wins", c.TargetWins, 1, 100); err != nil {
		return err
	}
	return nil
}
