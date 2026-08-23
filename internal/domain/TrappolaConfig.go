//go:build !js || !wasm || extra2

package domain

// TrappolaCpuDifficulty CPU の難易度レベル
type TrappolaCpuDifficulty int

// Trappola の CPU 難易度定数
const (
	// TrappolaCpuDifficultyEasy 低難易度（ランダムプレイ）
	TrappolaCpuDifficultyEasy TrappolaCpuDifficulty = iota
	// TrappolaCpuDifficultyNormal 中難易度（戦略プレイ）
	TrappolaCpuDifficultyNormal
	// TrappolaCpuDifficultyHard 高難易度（戦略プレイ）
	TrappolaCpuDifficultyHard
)

// TrappolaConfig トラッポラのゲーム設定
type TrappolaConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty TrappolaCpuDifficulty `json:"cd"`
	// TargetPoints ゲーム勝利に必要な得点。いずれかのチームの累積得点が
	// この値以上に達し、かつ相手チームを上回ったらゲーム終了。
	TargetPoints int `json:"tp"`
}

// DefaultTrappolaConfig デフォルト設定を返す（標準は 21 点先取）。
func DefaultTrappolaConfig() TrappolaConfig {
	return TrappolaConfig{CpuDifficulty: TrappolaCpuDifficultyNormal, TargetPoints: 21}
}

// Validate 設定値のドメインバリデーション
func (c TrappolaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TrappolaCpuDifficultyEasy), int(TrappolaCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
