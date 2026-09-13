//go:build !js || !wasm || extra3

package domain

// PochCpuDifficulty CPU の難易度レベル
type PochCpuDifficulty int

// Poch の CPU 難易度定数
const (
	// PochCpuDifficultyEasy 低難易度。
	PochCpuDifficultyEasy PochCpuDifficulty = iota
	// PochCpuDifficultyNormal 中難易度。
	PochCpuDifficultyNormal
	// PochCpuDifficultyHard 高難易度。
	PochCpuDifficultyHard
)

// PochConfig ポッホのゲーム設定
type PochConfig struct {
	CpuDifficulty PochCpuDifficulty `json:"cd"`
	// TargetDeals これだけディールを終えたら決着。
	TargetDeals int `json:"td"`
}

// DefaultPochConfig デフォルト設定を返す
func DefaultPochConfig() PochConfig {
	return PochConfig{
		CpuDifficulty: PochCpuDifficultyNormal,
		TargetDeals:   5,
	}
}

// Validate 設定値のドメインバリデーション
func (c PochConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(PochCpuDifficultyEasy), int(PochCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateRange("target deals", c.TargetDeals, 1, 100)
}
