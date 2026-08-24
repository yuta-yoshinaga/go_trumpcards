//go:build !js || !wasm || classic

package domain

// GermanSoloCpuDifficulty CPU の難易度レベル
type GermanSoloCpuDifficulty int

// GermanSolo の CPU 難易度定数
const (
	// GermanSoloCpuDifficultyEasy 低難易度 (ランダムプレイ・常にパス)
	GermanSoloCpuDifficultyEasy GermanSoloCpuDifficulty = iota
	// GermanSoloCpuDifficultyNormal 中難易度 (戦略プレイ)
	GermanSoloCpuDifficultyNormal
	// GermanSoloCpuDifficultyHard 高難易度 (戦略プレイ)
	GermanSoloCpuDifficultyHard
)

// GermanSoloConfig ジャーマン・ソロ (GermanSolo) のゲーム設定
type GermanSoloConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty GermanSoloCpuDifficulty `json:"cd"`
	// TargetRounds マッチを構成するディール数。この回数だけ配り、累積点最上位が勝者。
	TargetRounds int `json:"tr"`
}

// DefaultGermanSoloConfig デフォルト設定を返す (標準は 5 ディール)。
func DefaultGermanSoloConfig() GermanSoloConfig {
	return GermanSoloConfig{CpuDifficulty: GermanSoloCpuDifficultyNormal, TargetRounds: GermanSoloWinRounds}
}

// Validate 設定値のドメインバリデーション
func (c GermanSoloConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(GermanSoloCpuDifficultyEasy), int(GermanSoloCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
