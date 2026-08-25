//go:build !js || !wasm || extra2

package domain

import "fmt"

// ContinentalRummyCpuDifficulty は CPU の難易度。
type ContinentalRummyCpuDifficulty int

// CPU 難易度定数。
const (
	// ContinentalRummyCpuDifficultyEasy 低難易度 (捨てる札を雑に選ぶ)。
	ContinentalRummyCpuDifficultyEasy ContinentalRummyCpuDifficulty = iota
	// ContinentalRummyCpuDifficultyNormal 中難易度。
	ContinentalRummyCpuDifficultyNormal
	// ContinentalRummyCpuDifficultyHard 高難易度。
	ContinentalRummyCpuDifficultyHard
)

// ラウンド数の範囲。
const (
	// ContinentalRummyMinRounds は最少ラウンド数。
	ContinentalRummyMinRounds = 1
	// ContinentalRummyMaxRounds は最多ラウンド数。
	ContinentalRummyMaxRounds = 10
	// ContinentalRummyDefaultRounds は既定のラウンド数。
	ContinentalRummyDefaultRounds = 3
)

// ContinentalRummyRoundsOptions は選べるラウンド数。
var ContinentalRummyRoundsOptions = []int{1, 3, 5, 10}

// ContinentalRummyConfig はコンチネンタル・ラミーの設定。
type ContinentalRummyConfig struct {
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty ContinentalRummyCpuDifficulty `json:"cd"`
	// TotalRounds は打つラウンド数。
	TotalRounds int `json:"tr"`
}

// DefaultContinentalRummyConfig は既定の設定を返す。
func DefaultContinentalRummyConfig() ContinentalRummyConfig {
	return ContinentalRummyConfig{
		CpuDifficulty: ContinentalRummyCpuDifficultyNormal,
		TotalRounds:   ContinentalRummyDefaultRounds,
	}
}

// Validate は設定値を検査する。
func (c ContinentalRummyConfig) Validate() error {
	if c.CpuDifficulty < ContinentalRummyCpuDifficultyEasy || c.CpuDifficulty > ContinentalRummyCpuDifficultyHard {
		return NewDomainErrorCode(ErrInvalidAmount, "continentalrummy.invalidDifficulty",
			map[string]string{"val": fmt.Sprint(int(c.CpuDifficulty))})
	}
	if c.TotalRounds < ContinentalRummyMinRounds || c.TotalRounds > ContinentalRummyMaxRounds {
		return NewDomainErrorCode(ErrInvalidAmount, "continentalrummy.invalidRounds",
			map[string]string{"val": fmt.Sprint(c.TotalRounds)})
	}
	return nil
}
