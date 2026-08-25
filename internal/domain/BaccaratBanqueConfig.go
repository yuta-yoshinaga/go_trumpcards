//go:build !js || !wasm || extra2

package domain

// BaccaratBanqueCpuDifficulty は CPU の難易度。
type BaccaratBanqueCpuDifficulty int

// CPU 難易度定数。
const (
	// BaccaratBanqueCpuDifficultyEasy 低難易度 (裁量をランダムに決める)。
	BaccaratBanqueCpuDifficultyEasy BaccaratBanqueCpuDifficulty = iota
	// BaccaratBanqueCpuDifficultyNormal 中難易度 (慣習に沿って決める)。
	BaccaratBanqueCpuDifficultyNormal
	// BaccaratBanqueCpuDifficultyHard 高難易度 (慣習に沿って決める)。
	BaccaratBanqueCpuDifficultyHard
)

// チップ。
const (
	// BaccaratBanqueMinChips は持ちチップの下限。
	BaccaratBanqueMinChips = 100
	// BaccaratBanqueMaxChips は持ちチップの上限。
	BaccaratBanqueMaxChips = 100000
	// BaccaratBanqueDefaultChips は既定の持ちチップ。
	BaccaratBanqueDefaultChips = 1000
	// BaccaratBanqueMinBet は 1 つの子が張る額の下限。
	BaccaratBanqueMinBet = 10
	// BaccaratBanqueMaxBet は 1 つの子が張る額の上限。
	BaccaratBanqueMaxBet = 500
	// BaccaratBanqueDefaultBet は既定の張り額。
	BaccaratBanqueDefaultBet = 50
)

// BaccaratBanqueChipsOptions は選べる持ちチップ。
var BaccaratBanqueChipsOptions = []int{500, 1000, 5000}

// BaccaratBanqueBetOptions は選べる張り額。
var BaccaratBanqueBetOptions = []int{10, 50, 100, 500}

// BaccaratBanqueConfig はバカラ・バンクの設定。
type BaccaratBanqueConfig struct {
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty BaccaratBanqueCpuDifficulty `json:"cd"`
	// StartChips はバンカー (人間) の元手。
	StartChips int `json:"sc"`
	// BetAmount は 1 つの子が 1 回に張る額。
	BetAmount int `json:"ba"`
}

// DefaultBaccaratBanqueConfig は既定の設定を返す。
func DefaultBaccaratBanqueConfig() BaccaratBanqueConfig {
	return BaccaratBanqueConfig{
		CpuDifficulty: BaccaratBanqueCpuDifficultyNormal,
		StartChips:    BaccaratBanqueDefaultChips,
		BetAmount:     BaccaratBanqueDefaultBet,
	}
}

// Validate は設定値を検査する。
func (c BaccaratBanqueConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(BaccaratBanqueCpuDifficultyEasy), int(BaccaratBanqueCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("start chips", c.StartChips,
		BaccaratBanqueMinChips, BaccaratBanqueMaxChips); err != nil {
		return err
	}
	return ValidateRange("bet amount", c.BetAmount, BaccaratBanqueMinBet, BaccaratBanqueMaxBet)
}
