//go:build !js || !wasm || casino

package domain

// TuteCpuDifficulty CPU の難易度レベル
type TuteCpuDifficulty int

// Tute の CPU 難易度定数
const (
	// TuteCpuDifficultyEasy 低難易度（ランダムプレイ）
	TuteCpuDifficultyEasy TuteCpuDifficulty = iota
	// TuteCpuDifficultyNormal 中難易度（戦略プレイ）
	TuteCpuDifficultyNormal
	// TuteCpuDifficultyHard 高難易度（戦略プレイ）
	TuteCpuDifficultyHard
)

// TuteConfig トゥーテのゲーム設定
type TuteConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty TuteCpuDifficulty `json:"cd"`
	// TargetPoints ゲーム勝利に必要な累積点。いずれかのチームがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultTuteConfig デフォルト設定を返す（標準は 121 点先取）。
func DefaultTuteConfig() TuteConfig {
	return TuteConfig{CpuDifficulty: TuteCpuDifficultyNormal, TargetPoints: 121}
}

// Validate 設定値のドメインバリデーション
func (c TuteConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TuteCpuDifficultyEasy), int(TuteCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
