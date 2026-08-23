//go:build !js || !wasm || extra4

package domain

// QuadrilleCoalitionSize 連合 (非カドリール) 側の人数。得点移動の倍率に用いる。
const QuadrilleCoalitionSize = QuadrillePlayerCnt - 1

// QuadrilleCpuDifficulty CPU の難易度レベル
type QuadrilleCpuDifficulty int

// Quadrille の CPU 難易度定数
const (
	// QuadrilleCpuDifficultyEasy 低難易度 (ランダムプレイ・常にパス)
	QuadrilleCpuDifficultyEasy QuadrilleCpuDifficulty = iota
	// QuadrilleCpuDifficultyNormal 中難易度 (戦略プレイ)
	QuadrilleCpuDifficultyNormal
	// QuadrilleCpuDifficultyHard 高難易度 (戦略プレイ)
	QuadrilleCpuDifficultyHard
)

// QuadrilleConfig カドリール (Quadrille) のゲーム設定
type QuadrilleConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty QuadrilleCpuDifficulty `json:"cd"`
	// TargetRounds マッチを構成するディール数。この回数だけ配り、累積点最上位が勝者。
	TargetRounds int `json:"tr"`
}

// DefaultQuadrilleConfig デフォルト設定を返す (標準は 5 ディール)。
func DefaultQuadrilleConfig() QuadrilleConfig {
	return QuadrilleConfig{CpuDifficulty: QuadrilleCpuDifficultyNormal, TargetRounds: QuadrilleWinRounds}
}

// Validate 設定値のドメインバリデーション
func (c QuadrilleConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(QuadrilleCpuDifficultyEasy), int(QuadrilleCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
