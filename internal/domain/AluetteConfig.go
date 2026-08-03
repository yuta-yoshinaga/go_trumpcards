//go:build !js || !wasm || solo

package domain

// AluetteCpuDifficulty CPU の難易度。
type AluetteCpuDifficulty int

const (
	// AluetteCpuDifficultyEasy ランダムな合法手。
	AluetteCpuDifficultyEasy AluetteCpuDifficulty = iota
	// AluetteCpuDifficultyNormal 基本的な強弱判断。
	AluetteCpuDifficultyNormal
	// AluetteCpuDifficultyHard 発展的なヒューリスティック。
	AluetteCpuDifficultyHard
)

// AluetteConfig アリュエットのゲーム設定。
type AluetteConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty AluetteCpuDifficulty `json:"cd"`
	// TargetPoints マッチを取るのに必要なメーヌ数。
	//
	// **局数ではなく到達点。**アリュエットはメーヌを取った側に 1 点入り、先に
	// 規定点へ達した側の勝ち。ディーラーが一巡する必要は無いので、他ゲームのような
	// 「プレイヤー数の倍数」制約は課さない。
	TargetPoints int `json:"tp"`
}

// DefaultAluetteConfig 既定設定を返す。
func DefaultAluetteConfig() AluetteConfig {
	return AluetteConfig{CpuDifficulty: AluetteCpuDifficultyNormal, TargetPoints: 6}
}

// Validate 設定値を検証する。
func (c AluetteConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(AluetteCpuDifficultyEasy), int(AluetteCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateMin("target points", c.TargetPoints, 1)
}
