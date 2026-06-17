//go:build !js || !wasm || casino

package domain

// EcarteCpuDifficulty CPU の難易度レベル
type EcarteCpuDifficulty int

// Écarté の CPU 難易度定数
const (
	// EcarteCpuDifficultyEasy 低難易度
	EcarteCpuDifficultyEasy EcarteCpuDifficulty = iota
	// EcarteCpuDifficultyNormal 中難易度
	EcarteCpuDifficultyNormal
	// EcarteCpuDifficultyHard 高難易度
	EcarteCpuDifficultyHard
)

// EcarteDefaultTargetScore 試合終了スコア (先に到達した側が勝利)
const EcarteDefaultTargetScore = 5

// EcarteMaxTargetScore Validate で許容する TargetScore の上限
const EcarteMaxTargetScore = 50

// EcarteConfig Écarté ゲーム設定
type EcarteConfig struct {
	CpuDifficulty EcarteCpuDifficulty `json:"cd"`
	TargetScore   int                 `json:"ts"` // 試合終了スコア (デフォルト 5)
}

// DefaultEcarteConfig デフォルト設定を返す
func DefaultEcarteConfig() EcarteConfig {
	return EcarteConfig{
		CpuDifficulty: EcarteCpuDifficultyNormal,
		TargetScore:   EcarteDefaultTargetScore,
	}
}

// Validate 設定値のドメインバリデーション
func (c EcarteConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(EcarteCpuDifficultyEasy), int(EcarteCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("target score", c.TargetScore, 1, EcarteMaxTargetScore); err != nil {
		return err
	}
	return nil
}
