//go:build !js || !wasm || casino

package domain

// TeenPattiDefaultAnte デフォルトのアンティ額
const TeenPattiDefaultAnte = 1

// TeenPattiDefaultStartingChips デフォルトの初期チップ
const TeenPattiDefaultStartingChips = 30

// TeenPattiMaxStartingChips Validate で許容する初期チップ上限
const TeenPattiMaxStartingChips = 100000

// TeenPattiConfig Three Card Brag ゲーム設定
type TeenPattiConfig struct {
	Ante          int `json:"an"` // アンティ額 (デフォルト 1)
	StartingChips int `json:"sc"` // 初期チップ (デフォルト 30)
}

// DefaultTeenPattiConfig デフォルト設定を返す
func DefaultTeenPattiConfig() TeenPattiConfig {
	return TeenPattiConfig{
		Ante:          TeenPattiDefaultAnte,
		StartingChips: TeenPattiDefaultStartingChips,
	}
}

// Validate 設定値のドメインバリデーション
func (c TeenPattiConfig) Validate() error {
	if err := ValidateRange("ante", c.Ante, 1, 1000); err != nil {
		return err
	}
	if err := ValidateRange("starting chips", c.StartingChips, 2, TeenPattiMaxStartingChips); err != nil {
		return err
	}
	return nil
}
