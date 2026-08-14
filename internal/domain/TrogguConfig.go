//go:build !js || !wasm || extra

package domain

// TrogguCpuDifficulty CPU の難易度レベル
type TrogguCpuDifficulty int

// Troggu の CPU 難易度定数
const (
	// TrogguCpuDifficultyEasy 低難易度 (ランダムプレイ)
	TrogguCpuDifficultyEasy TrogguCpuDifficulty = iota
	// TrogguCpuDifficultyNormal 中難易度 (戦略プレイ)
	TrogguCpuDifficultyNormal
	// TrogguCpuDifficultyHard 高難易度 (戦略プレイ)
	TrogguCpuDifficultyHard
)

// TrogguConfig トロッグ (Troggu) のゲーム設定
type TrogguConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty TrogguCpuDifficulty `json:"cd"`
	// TargetDeals マッチを構成するディール数。この回数だけ配り、累積得点最上位が勝者。
	TargetDeals int `json:"td"`
}

// DefaultTrogguConfig デフォルト設定を返す (標準は 4 ディール)。
func DefaultTrogguConfig() TrogguConfig {
	return TrogguConfig{CpuDifficulty: TrogguCpuDifficultyNormal, TargetDeals: TrogguDefaultDeals}
}

// Validate 設定値のドメインバリデーション
func (c TrogguConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(TrogguCpuDifficultyEasy), int(TrogguCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateRange("target deals", c.TargetDeals, TrogguMinDeals, TrogguMaxDeals)
}
