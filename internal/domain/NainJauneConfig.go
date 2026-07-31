//go:build !js || !wasm || extra3

package domain

// NainJauneCpuDifficulty CPU の難易度レベル
type NainJauneCpuDifficulty int

// Nain Jaune の CPU 難易度定数
const (
	// NainJauneCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	NainJauneCpuDifficultyNormal NainJauneCpuDifficulty = iota
)

// NainJauneConfig ル・ナン・ジョーヌのゲーム設定
type NainJauneConfig struct {
	CpuDifficulty NainJauneCpuDifficulty `json:"cd"`
	// TargetDeals これだけディールを終えたら決着。
	TargetDeals int `json:"td"`
}

// DefaultNainJauneConfig デフォルト設定を返す
func DefaultNainJauneConfig() NainJauneConfig {
	return NainJauneConfig{
		CpuDifficulty: NainJauneCpuDifficultyNormal,
		TargetDeals:   5,
	}
}

// Validate 設定値のドメインバリデーション
func (c NainJauneConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(NainJauneCpuDifficultyNormal), int(NainJauneCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("target deals", c.TargetDeals, 1, 100)
}
