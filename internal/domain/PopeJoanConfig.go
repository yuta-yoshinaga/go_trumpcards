//go:build !js || !wasm || extra3

package domain

// PopeJoanCpuDifficulty CPU の難易度レベル
type PopeJoanCpuDifficulty int

// Pope Joan の CPU 難易度定数
const (
	// PopeJoanCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	PopeJoanCpuDifficultyNormal PopeJoanCpuDifficulty = iota
)

// PopeJoanConfig ポープ・ジョーンのゲーム設定
type PopeJoanConfig struct {
	CpuDifficulty PopeJoanCpuDifficulty `json:"cd"`
	// TargetDeals これだけディールを終えたら決着。
	TargetDeals int `json:"td"`
}

// DefaultPopeJoanConfig デフォルト設定を返す
func DefaultPopeJoanConfig() PopeJoanConfig {
	return PopeJoanConfig{
		CpuDifficulty: PopeJoanCpuDifficultyNormal,
		TargetDeals:   5,
	}
}

// Validate 設定値のドメインバリデーション
func (c PopeJoanConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(PopeJoanCpuDifficultyNormal), int(PopeJoanCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("target deals", c.TargetDeals, 1, 100)
}
