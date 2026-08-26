//go:build !js || !wasm || classic

package domain

// DilotiCpuDifficulty は CPU の難易度。
type DilotiCpuDifficulty int

// CPU 難易度定数。
const (
	// DilotiCpuDifficultyEasy 低難易度 (合法手からランダム)。
	DilotiCpuDifficultyEasy DilotiCpuDifficulty = iota
	// DilotiCpuDifficultyNormal 中難易度 (点になる札を狙う)。
	DilotiCpuDifficultyNormal
	// DilotiCpuDifficultyHard 高難易度 (点になる札を狙う)。
	DilotiCpuDifficultyHard
)

// 目標点。
const (
	// DilotiMinTarget は目標点の最小値。
	DilotiMinTarget = 21
	// DilotiMaxTarget は目標点の最大値。
	DilotiMaxTarget = 101
	// DilotiDefaultTarget は既定の目標点。**本来のディロティは 61 点勝負。**
	DilotiDefaultTarget = 61
)

// DilotiTargetOptions は選べる目標点。
var DilotiTargetOptions = []int{21, 41, 61, 101}

// DilotiConfig はディロティの設定。
type DilotiConfig struct {
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty DilotiCpuDifficulty `json:"cd"`
	// TargetScore は勝利に必要な点。
	TargetScore int `json:"ts"`
}

// DefaultDilotiConfig は既定の設定を返す。
func DefaultDilotiConfig() DilotiConfig {
	return DilotiConfig{
		CpuDifficulty: DilotiCpuDifficultyNormal,
		TargetScore:   DilotiDefaultTarget,
	}
}

// Validate は設定値を検査する。
func (c DilotiConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(DilotiCpuDifficultyEasy), int(DilotiCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, DilotiMinTarget, DilotiMaxTarget)
}
