//go:build !js || !wasm || classic

package domain

// SedmaCpuDifficulty CPU の難易度レベル
type SedmaCpuDifficulty int

// Sedma の CPU 難易度定数
const (
	// SedmaCpuDifficultyEasy 低難易度（ランダムプレイ）
	SedmaCpuDifficultyEasy SedmaCpuDifficulty = iota
	// SedmaCpuDifficultyNormal 中難易度（戦略プレイ）
	SedmaCpuDifficultyNormal
	// SedmaCpuDifficultyHard 高難易度（戦略プレイ）
	SedmaCpuDifficultyHard
)

// SedmaConfig セドマのゲーム設定
type SedmaConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty SedmaCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積点。いずれかのチームがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultSedmaConfig デフォルト設定を返す（標準は 101 点先取）。
func DefaultSedmaConfig() SedmaConfig {
	return SedmaConfig{CpuDifficulty: SedmaCpuDifficultyNormal, TargetPoints: 101}
}

// Validate 設定値のドメインバリデーション
func (c SedmaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SedmaCpuDifficultyEasy), int(SedmaCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
