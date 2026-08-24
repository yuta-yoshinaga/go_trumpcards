//go:build !js || !wasm || extra4

package domain

// PiedmonteseTarotCpuDifficulty は CPU の難易度レベル。
type PiedmonteseTarotCpuDifficulty int

// CPU 難易度定数。
const (
	// PiedmonteseTarotCpuDifficultyEasy 低難易度 (合法手からランダム)。
	PiedmonteseTarotCpuDifficultyEasy PiedmonteseTarotCpuDifficulty = iota
	// PiedmonteseTarotCpuDifficultyNormal 中難易度 (戦略プレイ)。
	PiedmonteseTarotCpuDifficultyNormal
	// PiedmonteseTarotCpuDifficultyHard 高難易度 (戦略プレイ)。
	PiedmonteseTarotCpuDifficultyHard
)

// 卓の大きさ。
const (
	// PiedmonteseTarotMinSeats は最小席数。
	PiedmonteseTarotMinSeats = 3
	// PiedmonteseTarotMaxSeats は最大席数。
	PiedmonteseTarotMaxSeats = 4
	// PiedmonteseTarotDefaultSeats は既定の席数 (標準は 4 人)。
	PiedmonteseTarotDefaultSeats = 4
	// PiedmonteseTarotDefaultDeals は既定のディール数。
	PiedmonteseTarotDefaultDeals = 4
)

// PiedmonteseTarotSeatSizes は選べる席数。
var PiedmonteseTarotSeatSizes = []int{3, 4}

// PiedmonteseTarotConfig はピエモンテ・タロッコのゲーム設定。
type PiedmonteseTarotConfig struct {
	// Seats は卓の人数 (3 または 4)。**配り方が変わる** ── 4 人なら 19 枚 + タロン 2、
	// 3 人なら 25 枚 + タロン 3。
	Seats int `json:"s"`
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty PiedmonteseTarotCpuDifficulty `json:"cd"`
	// TargetDeals はマッチを構成するディール数。この回数だけ配り、累積得点最上位が勝者。
	TargetDeals int `json:"td"`
}

// DefaultPiedmonteseTarotConfig は既定設定を返す。
func DefaultPiedmonteseTarotConfig() PiedmonteseTarotConfig {
	return PiedmonteseTarotConfig{
		Seats:         PiedmonteseTarotDefaultSeats,
		CpuDifficulty: PiedmonteseTarotCpuDifficultyNormal,
		TargetDeals:   PiedmonteseTarotDefaultDeals,
	}
}

// Validate は設定値のドメインバリデーション。
func (c PiedmonteseTarotConfig) Validate() error {
	if err := ValidateRange("seats", c.Seats, PiedmonteseTarotMinSeats, PiedmonteseTarotMaxSeats); err != nil {
		return err
	}
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(PiedmonteseTarotCpuDifficultyEasy), int(PiedmonteseTarotCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target deals", c.TargetDeals, 1); err != nil {
		return err
	}
	return nil
}

// piedmonteseTarotHandSizes は席数ごとの 1 人の手札枚数。
//
// **割り切れる数を配るのではない。** 78 ÷ 3 = 26 はちょうど割り切れてしまい、
// そのまま配ると**タロンが 0 枚**になってスカルト (親の捨て札) が空回りする ──
// 実際に 3 人卓を 26 枚配りで回して、捨てる札が 1 枚も無いまま「捨てた」ことに
// なる盤面を作ってしまった。ピエモンテの配りは 3 人なら 25 枚 + タロン 3 枚、
// 4 人なら 19 枚 + タロン 2 枚で、**タロンは規則が決めている**。
var piedmonteseTarotHandSizes = map[int]int{
	3: 25,
	4: 19,
}

// PiedmonteseTarotHandSize は席数に応じた 1 人の手札枚数を返す (不正な席数は 0)。
func PiedmonteseTarotHandSize(seats int) int {
	return piedmonteseTarotHandSizes[seats]
}

// PiedmonteseTarotTalonSize は席数に応じたタロン (親が拾って捨てる余剰札) の枚数を返す。
func PiedmonteseTarotTalonSize(seats int) int {
	hand := PiedmonteseTarotHandSize(seats)
	if hand == 0 {
		return 0
	}
	return Tarot78DeckSize - hand*seats
}
