//go:build !js || !wasm || extra4

package domain

// CegoCpuDifficulty CPU の難易度レベル
type CegoCpuDifficulty int

// Cego の CPU 難易度定数
const (
	// CegoCpuDifficultyEasy 低難易度 (ランダムプレイ)
	CegoCpuDifficultyEasy CegoCpuDifficulty = iota
	// CegoCpuDifficultyNormal 中難易度 (戦略プレイ)
	CegoCpuDifficultyNormal
	// CegoCpuDifficultyHard 高難易度 (戦略プレイ)
	CegoCpuDifficultyHard
)

// CegoConfig チェゴ (Cego) のゲーム設定
type CegoConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty CegoCpuDifficulty `json:"cd"`
	// TargetDeals マッチを構成するディール数。この回数だけ配り、累積得点最上位が勝者。
	TargetDeals int `json:"td"`
}

// DefaultCegoConfig デフォルト設定を返す (標準は 5 ディール)。
func DefaultCegoConfig() CegoConfig {
	return CegoConfig{CpuDifficulty: CegoCpuDifficultyNormal, TargetDeals: CegoDefaultDeals}
}

// Validate 設定値のドメインバリデーション
func (c CegoConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CegoCpuDifficultyEasy), int(CegoCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target deals", c.TargetDeals, 1); err != nil {
		return err
	}
	return nil
}
