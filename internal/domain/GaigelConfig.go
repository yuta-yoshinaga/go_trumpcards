//go:build !js || !wasm || extra

package domain

// GaigelCpuDifficulty CPU の難易度レベル
type GaigelCpuDifficulty int

// GaigelのCPU難易度定数
const (
	// GaigelCpuDifficultyEasy 低難易度
	GaigelCpuDifficultyEasy GaigelCpuDifficulty = iota
	// GaigelCpuDifficultyNormal 中難易度
	GaigelCpuDifficultyNormal
	// GaigelCpuDifficultyHard 高難易度
	GaigelCpuDifficultyHard
)

// GaigelConfig ガイゲルゲーム設定
type GaigelConfig struct {
	CpuDifficulty GaigelCpuDifficulty `json:"cd"`
	TargetScore   int                 `json:"ts"` // ゲーム終了スコア (先に到達したチームが勝利, デフォルト101)
}

// DefaultGaigelConfig デフォルト設定を返す
func DefaultGaigelConfig() GaigelConfig {
	return GaigelConfig{
		CpuDifficulty: GaigelCpuDifficultyNormal,
		TargetScore:   101,
	}
}

// Validate 設定値のドメインバリデーション
func (c GaigelConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(GaigelCpuDifficultyEasy), int(GaigelCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target score", c.TargetScore, 1); err != nil {
		return err
	}
	return nil
}
