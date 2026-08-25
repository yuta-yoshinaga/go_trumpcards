//go:build !js || !wasm || solo

package domain

// CometCpuDifficulty は CPU の難易度。
type CometCpuDifficulty int

// CPU 難易度定数。
const (
	// CometCpuDifficultyEasy 低難易度 (合法手からランダム)。
	CometCpuDifficultyEasy CometCpuDifficulty = iota
	// CometCpuDifficultyNormal 中難易度 (コメットと K を溜める)。
	CometCpuDifficultyNormal
	// CometCpuDifficultyHard 高難易度 (コメットと K を溜める)。
	CometCpuDifficultyHard
)

// 席数。
const (
	// CometMinPlayers は最小の席数。
	CometMinPlayers = 2
	// CometMaxPlayers は最大の席数。**5 人までで頭打ち。**
	CometMaxPlayers = 5
	// CometDefaultPlayers は既定の席数。
	CometDefaultPlayers = 4
)

// 目標点。
const (
	// CometMinTarget は目標点の最小値。
	CometMinTarget = 20
	// CometMaxTarget は目標点の最大値。
	CometMaxTarget = 200
	// CometDefaultTarget は既定の目標点。**本来のコメットは 100 点勝負。**
	CometDefaultTarget = 100
)

// CometTargetOptions は選べる目標点。
var CometTargetOptions = []int{20, 50, 100, 200}

// CometConfig はコメットの設定。
type CometConfig struct {
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty CometCpuDifficulty `json:"cd"`
	// Players は席数 (人間 1 + CPU)。
	Players int `json:"pl"`
	// TargetScore は勝利に必要な点。
	TargetScore int `json:"ts"`
}

// DefaultCometConfig は既定の設定を返す。
func DefaultCometConfig() CometConfig {
	return CometConfig{
		CpuDifficulty: CometCpuDifficultyNormal,
		Players:       CometDefaultPlayers,
		TargetScore:   CometDefaultTarget,
	}
}

// Validate は設定値を検査する。
func (c CometConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(CometCpuDifficultyEasy), int(CometCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("player count", c.Players, CometMinPlayers, CometMaxPlayers); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, CometMinTarget, CometMaxTarget)
}
