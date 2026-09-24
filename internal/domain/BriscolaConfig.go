package domain

// BriscolaConfig ブリスコラゲーム設定
type BriscolaConfig struct{}

// DefaultBriscolaConfig デフォルト設定を返す
func DefaultBriscolaConfig() BriscolaConfig {
	return BriscolaConfig{}
}

// Validate 設定値のドメインバリデーション
func (c BriscolaConfig) Validate() error {
	return nil
}
