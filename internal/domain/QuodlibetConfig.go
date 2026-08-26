//go:build !js || !wasm || solo

package domain

// QuodlibetCpuDifficulty は CPU の難易度。
type QuodlibetCpuDifficulty int

// CPU 難易度定数。
const (
	// QuodlibetCpuDifficultyEasy 低難易度 (合法手からランダム)。
	QuodlibetCpuDifficultyEasy QuodlibetCpuDifficulty = iota
	// QuodlibetCpuDifficultyNormal 中難易度 (罰点を避ける)。
	QuodlibetCpuDifficultyNormal
	// QuodlibetCpuDifficultyHard 高難易度 (罰点を避ける)。
	QuodlibetCpuDifficultyHard
)

// QuodlibetConfig はクオドリベットの設定。
type QuodlibetConfig struct {
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty QuodlibetCpuDifficulty `json:"cd"`
	// AutoSelectContract はディーラーが人間のときもコントラクトを自動で選ぶ。
	//
	// **12 ディールすべてで選択を求めると手数が増えるだけ** の卓もあるので、
	// 「輪の中の残り物を順に消化する」既定の遊び方を選べるようにしてある。
	AutoSelectContract bool `json:"as"`
}

// DefaultQuodlibetConfig は既定の設定を返す。
func DefaultQuodlibetConfig() QuodlibetConfig {
	return QuodlibetConfig{
		CpuDifficulty:      QuodlibetCpuDifficultyNormal,
		AutoSelectContract: false,
	}
}

// Validate は設定値を検査する。
func (c QuodlibetConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(QuodlibetCpuDifficultyEasy), int(QuodlibetCpuDifficultyHard))
}
