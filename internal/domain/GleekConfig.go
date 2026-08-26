//go:build !js || !wasm || extra

package domain

// GleekCpuDifficulty CPU の難易度レベル
type GleekCpuDifficulty int

// Gleek の CPU 難易度定数
const (
	// GleekCpuDifficultyEasy 低難易度 (ランダムプレイ・最低限の競り)
	GleekCpuDifficultyEasy GleekCpuDifficulty = iota
	// GleekCpuDifficultyNormal 中難易度 (戦略プレイ)
	GleekCpuDifficultyNormal
	// GleekCpuDifficultyHard 高難易度 (戦略プレイ)
	GleekCpuDifficultyHard
)

// GleekConfig グリーク (Gleek) のゲーム設定
type GleekConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty GleekCpuDifficulty `json:"cd"`
	// TargetRounds マッチを構成するディール数。この回数だけ配り、累積点最上位が勝者。
	TargetRounds int `json:"tr"`
}

// DefaultGleekConfig デフォルト設定を返す (標準は 5 ディール)。
func DefaultGleekConfig() GleekConfig {
	return GleekConfig{CpuDifficulty: GleekCpuDifficultyNormal, TargetRounds: GleekWinRounds}
}

// Validate 設定値のドメインバリデーション
func (c GleekConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(GleekCpuDifficultyEasy), int(GleekCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
