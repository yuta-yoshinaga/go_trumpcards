//go:build !js || !wasm || extra2

package domain

// SutdaCpuDifficulty は CPU の難易度。
type SutdaCpuDifficulty int

// CPU 難易度定数。
const (
	// SutdaCpuDifficultyEasy 低難易度 (ほぼコール)。
	SutdaCpuDifficultyEasy SutdaCpuDifficulty = iota
	// SutdaCpuDifficultyNormal 中難易度 (役の強さで判断)。
	SutdaCpuDifficultyNormal
	// SutdaCpuDifficultyHard 高難易度 (役の強さで判断)。
	SutdaCpuDifficultyHard
)

// 卓の人数。
const (
	// SutdaMinSeats は最小の席数。
	SutdaMinSeats = 2
	// SutdaMaxSeats は最大の席数。**20 枚を 2 枚ずつ配るので 10 席が上限**だが、
	// 1 人プレイの卓としては 5 席までにしてある。
	SutdaMaxSeats = 5
	// SutdaDefaultSeats は既定の席数。
	SutdaDefaultSeats = 3
)

// チップ。
const (
	// SutdaMinChips は開始チップの最小値。
	SutdaMinChips = 100
	// SutdaMaxChips は開始チップの最大値。
	SutdaMaxChips = 10000
	// SutdaDefaultChips は既定の開始チップ。
	SutdaDefaultChips = 1000
	// SutdaAnte は毎ハンドの参加料。
	SutdaAnte = 10
	// SutdaBetUnit は 1 回のベット / レイズの単位。
	SutdaBetUnit = 20
	// SutdaMaxRaises は 1 ハンドのレイズ上限。**上限が無いと CPU 同士が
	// 際限なく上げ続ける。**
	SutdaMaxRaises = 3
)

// SutdaSeatOptions は選べる席数。
var SutdaSeatOptions = []int{2, 3, 4, 5}

// SutdaConfig はソッタの設定。
type SutdaConfig struct {
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty SutdaCpuDifficulty `json:"cd"`
	// Seats は席数 (人間 1 + CPU)。
	Seats int `json:"st"`
	// StartChips は開始チップ。
	StartChips int `json:"sc"`
}

// DefaultSutdaConfig は既定の設定を返す。
func DefaultSutdaConfig() SutdaConfig {
	return SutdaConfig{
		CpuDifficulty: SutdaCpuDifficultyNormal,
		Seats:         SutdaDefaultSeats,
		StartChips:    SutdaDefaultChips,
	}
}

// Validate は設定値を検査する。
func (c SutdaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(SutdaCpuDifficultyEasy), int(SutdaCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("seats", c.Seats, SutdaMinSeats, SutdaMaxSeats); err != nil {
		return err
	}
	return ValidateRange("start chips", c.StartChips, SutdaMinChips, SutdaMaxChips)
}
