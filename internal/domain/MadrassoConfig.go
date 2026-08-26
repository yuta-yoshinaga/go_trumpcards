//go:build !js || !wasm || extra3

package domain

// MadrassoCpuDifficulty CPU の難易度レベル
type MadrassoCpuDifficulty int

// Madrasso の CPU 難易度定数
const (
	// MadrassoCpuDifficultyEasy 低難易度（ランダムプレイ）
	MadrassoCpuDifficultyEasy MadrassoCpuDifficulty = iota
	// MadrassoCpuDifficultyNormal 中難易度（戦略プレイ）
	MadrassoCpuDifficultyNormal
	// MadrassoCpuDifficultyHard 高難易度（戦略プレイ）
	MadrassoCpuDifficultyHard
)

// MadrassoConfig マドラッソのゲーム設定
type MadrassoConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty MadrassoCpuDifficulty `json:"cd"`
	// TargetPoints ゲーム勝利に必要な得点。いずれかのチームの累積得点が
	// この値以上に達し、かつ相手チームを上回ったらゲーム終了。
	TargetPoints int `json:"tp"`
}

// DefaultMadrassoConfig デフォルト設定を返す（標準は 21 点先取）。
func DefaultMadrassoConfig() MadrassoConfig {
	return MadrassoConfig{CpuDifficulty: MadrassoCpuDifficultyNormal, TargetPoints: 4}
}

// Validate 設定値のドメインバリデーション
func (c MadrassoConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(MadrassoCpuDifficultyEasy), int(MadrassoCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
