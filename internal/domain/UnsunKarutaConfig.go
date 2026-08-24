//go:build !js || !wasm || classic

package domain

// UnsunKarutaCpuDifficulty は CPU の難易度。
type UnsunKarutaCpuDifficulty int

// CPU 難易度定数。
const (
	// UnsunKarutaCpuDifficultyEasy 低難易度 (合法手からランダム)。
	UnsunKarutaCpuDifficultyEasy UnsunKarutaCpuDifficulty = iota
	// UnsunKarutaCpuDifficultyNormal 中難易度 (戦略プレイ)。
	UnsunKarutaCpuDifficultyNormal
	// UnsunKarutaCpuDifficultyHard 高難易度 (戦略プレイ)。
	UnsunKarutaCpuDifficultyHard
)

// マッチの長さ。
const (
	// UnsunKarutaMinDeals は最小のディール数。
	UnsunKarutaMinDeals = 1
	// UnsunKarutaMaxDeals は最大のディール数。**8 人が 1 回ずつ親を務めて 1 巡。**
	UnsunKarutaMaxDeals = 8
	// UnsunKarutaDefaultDeals は既定のディール数。
	UnsunKarutaDefaultDeals = 4
)

// UnsunKarutaDealOptions は選べるディール数。
var UnsunKarutaDealOptions = []int{1, 2, 4, 8}

// UnsunKarutaConfig はうんすんカルタ (八人メリ) の設定。
type UnsunKarutaConfig struct {
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty UnsunKarutaCpuDifficulty `json:"cd"`
	// TargetDeals はマッチを構成するディール数。多く「コ」を取ったチームの勝ち。
	TargetDeals int `json:"td"`
}

// DefaultUnsunKarutaConfig は既定の設定を返す。
func DefaultUnsunKarutaConfig() UnsunKarutaConfig {
	return UnsunKarutaConfig{
		CpuDifficulty: UnsunKarutaCpuDifficultyNormal,
		TargetDeals:   UnsunKarutaDefaultDeals,
	}
}

// Validate は設定値を検査する。
func (c UnsunKarutaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(UnsunKarutaCpuDifficultyEasy), int(UnsunKarutaCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateRange("target deals", c.TargetDeals, UnsunKarutaMinDeals, UnsunKarutaMaxDeals)
}
