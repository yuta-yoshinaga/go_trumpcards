//go:build !js || !wasm || extra3

package domain

// RookCpuDifficulty CPU の難易度レベル
type RookCpuDifficulty int

// Rook の CPU 難易度定数
const (
	// RookCpuDifficultyEasy 低難易度
	RookCpuDifficultyEasy RookCpuDifficulty = iota
	// RookCpuDifficultyNormal 中難易度
	RookCpuDifficultyNormal
	// RookCpuDifficultyHard 高難易度
	RookCpuDifficultyHard
)

// RookDefaultTargetScore ゲーム終了スコア (先に到達したチームが勝利)
const RookDefaultTargetScore = 500

// RookConfig ルーク(Rook)ゲーム設定
type RookConfig struct {
	CpuDifficulty RookCpuDifficulty `json:"cd"`
	TargetScore   int               `json:"ts"` // 勝利スコア (デフォルト500)
}

// DefaultRookConfig デフォルト設定を返す
func DefaultRookConfig() RookConfig {
	return RookConfig{
		CpuDifficulty: RookCpuDifficultyNormal,
		TargetScore:   RookDefaultTargetScore,
	}
}

// Validate 設定値のドメインバリデーション
func (c RookConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(RookCpuDifficultyEasy), int(RookCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target score", c.TargetScore, 1); err != nil {
		return err
	}
	return nil
}
