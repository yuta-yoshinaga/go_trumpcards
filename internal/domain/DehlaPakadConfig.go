//go:build !js || !wasm || extra

package domain

// DehlaPakadCpuDifficulty は CPU の難易度。
type DehlaPakadCpuDifficulty int

// CPU 難易度定数。
const (
	// DehlaPakadCpuDifficultyEasy 低難易度 (合法手からランダム)。
	DehlaPakadCpuDifficultyEasy DehlaPakadCpuDifficulty = iota
	// DehlaPakadCpuDifficultyNormal 中難易度 (10 を狙う / 守る)。
	DehlaPakadCpuDifficultyNormal
	// DehlaPakadCpuDifficultyHard 高難易度 (10 を狙う / 守る)。
	DehlaPakadCpuDifficultyHard
)

// マッチの長さ。
const (
	// DehlaPakadMinKots は勝利に必要なコートの最小値。
	DehlaPakadMinKots = 1
	// DehlaPakadMaxKots は勝利に必要なコートの最大値。
	DehlaPakadMaxKots = 5
	// DehlaPakadDefaultKots は既定のコート数。
	DehlaPakadDefaultKots = 2
)

// DehlaPakadKotOptions は選べるコート数。
var DehlaPakadKotOptions = []int{1, 2, 3, 5}

// DehlaPakadConfig はデーラ・パカドの設定。
type DehlaPakadConfig struct {
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty DehlaPakadCpuDifficulty `json:"cd"`
	// TargetKots は勝利に必要なコート数。
	TargetKots int `json:"tk"`
}

// DefaultDehlaPakadConfig は既定の設定を返す。
func DefaultDehlaPakadConfig() DehlaPakadConfig {
	return DehlaPakadConfig{
		CpuDifficulty: DehlaPakadCpuDifficultyNormal,
		TargetKots:    DehlaPakadDefaultKots,
	}
}

// Validate は設定値を検査する。
func (c DehlaPakadConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(DehlaPakadCpuDifficultyEasy), int(DehlaPakadCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateRange("target kots", c.TargetKots, DehlaPakadMinKots, DehlaPakadMaxKots)
}
