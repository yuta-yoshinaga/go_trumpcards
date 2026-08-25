//go:build !js || !wasm || extra

package domain

// CostlyColoursCpuDifficulty は CPU の難易度。
type CostlyColoursCpuDifficulty int

// CPU 難易度定数。
const (
	// CostlyColoursCpuDifficultyEasy 低難易度 (合法手からランダム)。
	CostlyColoursCpuDifficultyEasy CostlyColoursCpuDifficulty = iota
	// CostlyColoursCpuDifficultyNormal 中難易度 (節目を狙う)。
	CostlyColoursCpuDifficultyNormal
	// CostlyColoursCpuDifficultyHard 高難易度 (節目を狙う)。
	CostlyColoursCpuDifficultyHard
)

// 目標点。
//
// **出典が割れている。** Cotton (1674) は 61 点、Parlett は 121 点と書く。
// どちらも実在する遊び方なので選べるようにし、#5461 と Cotton の原典に
// 合わせて **61 を既定**にする。
const (
	// CostlyColoursMinTarget は目標点の最小値。
	CostlyColoursMinTarget = 31
	// CostlyColoursMaxTarget は目標点の最大値 (Parlett 版)。
	CostlyColoursMaxTarget = 121
	// CostlyColoursDefaultTarget は既定の目標点 (Cotton 版)。
	CostlyColoursDefaultTarget = 61
)

// CostlyColoursTargetOptions は選べる目標点。
var CostlyColoursTargetOptions = []int{31, 61, 121}

// CostlyColoursConfig はコストリー・カラーズの設定。
type CostlyColoursConfig struct {
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty CostlyColoursCpuDifficulty `json:"cd"`
	// TargetScore は勝利に必要な点。
	TargetScore int `json:"ts"`
}

// DefaultCostlyColoursConfig は既定の設定を返す。
func DefaultCostlyColoursConfig() CostlyColoursConfig {
	return CostlyColoursConfig{
		CpuDifficulty: CostlyColoursCpuDifficultyNormal,
		TargetScore:   CostlyColoursDefaultTarget,
	}
}

// Validate は設定値を検査する。
func (c CostlyColoursConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(CostlyColoursCpuDifficultyEasy), int(CostlyColoursCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore,
		CostlyColoursMinTarget, CostlyColoursMaxTarget)
}
