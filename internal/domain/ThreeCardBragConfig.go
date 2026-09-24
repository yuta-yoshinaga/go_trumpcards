//go:build !js || !wasm || casino

package domain

// ThreeCardBragDefaultAnte デフォルトのアンティ額
const ThreeCardBragDefaultAnte = 1

// ThreeCardBragDefaultStartingChips デフォルトの初期チップ
const ThreeCardBragDefaultStartingChips = 30

// ThreeCardBragMaxStartingChips Validate で許容する初期チップ上限
const ThreeCardBragMaxStartingChips = 100000

// ThreeCardBragConfig Three Card Brag ゲーム設定
type ThreeCardBragConfig struct {
	Ante          int `json:"an"` // アンティ額 (デフォルト 1)
	StartingChips int `json:"sc"` // 初期チップ (デフォルト 30)
}

// DefaultThreeCardBragConfig デフォルト設定を返す
func DefaultThreeCardBragConfig() ThreeCardBragConfig {
	return ThreeCardBragConfig{
		Ante:          ThreeCardBragDefaultAnte,
		StartingChips: ThreeCardBragDefaultStartingChips,
	}
}

// Validate 設定値のドメインバリデーション
func (c ThreeCardBragConfig) Validate() error {
	if err := ValidateRange("ante", c.Ante, 1, 1000); err != nil {
		return err
	}
	if err := ValidateRange("starting chips", c.StartingChips, 2, ThreeCardBragMaxStartingChips); err != nil {
		return err
	}
	return nil
}
