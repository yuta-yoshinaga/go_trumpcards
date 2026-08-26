//go:build !js || !wasm || extra4

package domain

// BeziqueCpuDifficulty CPU の難易度レベル
type BeziqueCpuDifficulty int

// Bezique の CPU 難易度定数
const (
	// BeziqueCpuDifficultyEasy 低難易度
	BeziqueCpuDifficultyEasy BeziqueCpuDifficulty = iota
	// BeziqueCpuDifficultyNormal 中難易度
	BeziqueCpuDifficultyNormal
	// BeziqueCpuDifficultyHard 高難易度
	BeziqueCpuDifficultyHard
)

// BeziqueDefaultTargetScore 試合終了スコア (先に到達した側が勝利)。
// 古典ベジークの 1500 点は簡略化のため既定 1000 点とする (設定で変更可)。
const BeziqueDefaultTargetScore = 1000

// BeziqueMaxTargetScore Validate で許容する TargetScore の上限
const BeziqueMaxTargetScore = 5000

// BeziqueConfig Bezique ゲーム設定
type BeziqueConfig struct {
	CpuDifficulty BeziqueCpuDifficulty `json:"cd"`
	TargetScore   int                  `json:"ts"` // 試合終了スコア (デフォルト 1000)
}

// DefaultBeziqueConfig デフォルト設定を返す
func DefaultBeziqueConfig() BeziqueConfig {
	return BeziqueConfig{
		CpuDifficulty: BeziqueCpuDifficultyNormal,
		TargetScore:   BeziqueDefaultTargetScore,
	}
}

// Validate 設定値のドメインバリデーション
func (c BeziqueConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(BeziqueCpuDifficultyEasy), int(BeziqueCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("target score", c.TargetScore, 100, BeziqueMaxTargetScore); err != nil {
		return err
	}
	return nil
}
