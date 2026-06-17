//go:build !js || !wasm || casino

package domain

// ThreeCardBragCpuDifficulty CPU の難易度レベル
type ThreeCardBragCpuDifficulty int

// Three Card Brag の CPU 難易度定数
const (
	// ThreeCardBragCpuDifficultyEasy 低難易度
	ThreeCardBragCpuDifficultyEasy ThreeCardBragCpuDifficulty = iota
	// ThreeCardBragCpuDifficultyNormal 中難易度
	ThreeCardBragCpuDifficultyNormal
	// ThreeCardBragCpuDifficultyHard 高難易度
	ThreeCardBragCpuDifficultyHard
)

// ThreeCardBragDefaultAnte デフォルトのアンティ額
const ThreeCardBragDefaultAnte = 1

// ThreeCardBragDefaultStartingChips デフォルトの初期チップ
const ThreeCardBragDefaultStartingChips = 30

// ThreeCardBragMaxStartingChips Validate で許容する初期チップ上限
const ThreeCardBragMaxStartingChips = 100000

// ThreeCardBragConfig Three Card Brag ゲーム設定
type ThreeCardBragConfig struct {
	CpuDifficulty ThreeCardBragCpuDifficulty `json:"cd"`
	Ante          int                        `json:"an"` // アンティ額 (デフォルト 1)
	StartingChips int                        `json:"sc"` // 初期チップ (デフォルト 30)
}

// DefaultThreeCardBragConfig デフォルト設定を返す
func DefaultThreeCardBragConfig() ThreeCardBragConfig {
	return ThreeCardBragConfig{
		CpuDifficulty: ThreeCardBragCpuDifficultyNormal,
		Ante:          ThreeCardBragDefaultAnte,
		StartingChips: ThreeCardBragDefaultStartingChips,
	}
}

// Validate 設定値のドメインバリデーション
func (c ThreeCardBragConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ThreeCardBragCpuDifficultyEasy), int(ThreeCardBragCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("ante", c.Ante, 1, 1000); err != nil {
		return err
	}
	if err := ValidateRange("starting chips", c.StartingChips, 2, ThreeCardBragMaxStartingChips); err != nil {
		return err
	}
	return nil
}
