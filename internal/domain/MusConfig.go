//go:build !js || !wasm || solo

package domain

// MusCpuDifficulty CPU の難易度レベル
type MusCpuDifficulty int

// Mus の CPU 難易度定数
const (
	// MusCpuDifficultyEasy 低難易度（保守的・ランダム寄り）
	MusCpuDifficultyEasy MusCpuDifficulty = iota
	// MusCpuDifficultyNormal 中難易度（手の強さに応じた賭け）
	MusCpuDifficultyNormal
	// MusCpuDifficultyHard 高難易度（手の強さに応じた賭け）
	MusCpuDifficultyHard
)

// MusConfig ムスのゲーム設定
type MusConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty MusCpuDifficulty `json:"cd"`
	// TargetAmarrakos ゲーム勝利に必要な点（アマ）。いずれかのチームがこの値以上で勝利。
	TargetAmarrakos int `json:"ta"`
}

// DefaultMusConfig デフォルト設定を返す（標準は 40 点先取）。
func DefaultMusConfig() MusConfig {
	return MusConfig{CpuDifficulty: MusCpuDifficultyNormal, TargetAmarrakos: 40}
}

// Validate 設定値のドメインバリデーション
func (c MusConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(MusCpuDifficultyEasy), int(MusCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target amarrakos", c.TargetAmarrakos, 1); err != nil {
		return err
	}
	return nil
}
