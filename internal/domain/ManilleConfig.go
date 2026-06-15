//go:build !js || !wasm || classic

package domain

// ManilleCpuDifficulty CPU の難易度レベル
type ManilleCpuDifficulty int

// Manille の CPU 難易度定数
const (
	// ManilleCpuDifficultyEasy 低難易度（ランダムプレイ）
	ManilleCpuDifficultyEasy ManilleCpuDifficulty = iota
	// ManilleCpuDifficultyNormal 中難易度（戦略プレイ）
	ManilleCpuDifficultyNormal
	// ManilleCpuDifficultyHard 高難易度（戦略プレイ）
	ManilleCpuDifficultyHard
)

// ManilleConfig マニーユのゲーム設定
type ManilleConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty ManilleCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積点。いずれかのチームがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultManilleConfig デフォルト設定を返す（標準は 101 点先取）。
func DefaultManilleConfig() ManilleConfig {
	return ManilleConfig{CpuDifficulty: ManilleCpuDifficultyNormal, TargetPoints: 101}
}

// Validate 設定値のドメインバリデーション
func (c ManilleConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ManilleCpuDifficultyEasy), int(ManilleCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
