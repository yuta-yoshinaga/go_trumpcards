//go:build !js || !wasm || extra3

package domain

// CirullaCpuDifficulty は CPU の難易度。
type CirullaCpuDifficulty int

// CPU 難易度定数。
const (
	// CirullaCpuDifficultyEasy 低難易度 (合法手からランダム)。
	CirullaCpuDifficultyEasy CirullaCpuDifficulty = iota
	// CirullaCpuDifficultyNormal 中難易度 (点になる札を狙う)。
	CirullaCpuDifficultyNormal
	// CirullaCpuDifficultyHard 高難易度 (点になる札を狙う)。
	CirullaCpuDifficultyHard
)

// 目標点。
const (
	// CirullaMinTarget は目標点の最小値。
	CirullaMinTarget = 11
	// CirullaMaxTarget は目標点の最大値。
	CirullaMaxTarget = 51
	// CirullaDefaultTarget は既定の目標点。**本来のチルッラは 51 点勝負。**
	CirullaDefaultTarget = 51
)

// CirullaTargetOptions は選べる目標点。
var CirullaTargetOptions = []int{11, 21, 31, 51}

// CirullaConfig はチルッラの設定。
type CirullaConfig struct {
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty CirullaCpuDifficulty `json:"cd"`
	// TargetScore は勝利に必要な点。
	TargetScore int `json:"ts"`
}

// DefaultCirullaConfig は既定の設定を返す。
func DefaultCirullaConfig() CirullaConfig {
	return CirullaConfig{
		CpuDifficulty: CirullaCpuDifficultyNormal,
		TargetScore:   CirullaDefaultTarget,
	}
}

// Validate は設定値を検査する。
func (c CirullaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(CirullaCpuDifficultyEasy), int(CirullaCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, CirullaMinTarget, CirullaMaxTarget)
}
