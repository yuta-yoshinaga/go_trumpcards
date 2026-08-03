//go:build !js || !wasm || extra2

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

// DefaultAluetteTargetPoints 既定の到達点。
//
// **画面・マニュアル・フロントの既定と同じ値であること。**ここだけ動かすと
// CUI は 6 点先取なのにマニュアルは 5 点先取、という食い違いが黙って生まれる
// (PR #4666 のレビューで実際に指摘された)。
const DefaultAluetteTargetPoints = 5

// DefaultAluetteConfig 既定設定を返す。
func DefaultAluetteConfig() AluetteConfig {
	return AluetteConfig{CpuDifficulty: AluetteCpuDifficultyNormal, TargetPoints: DefaultAluetteTargetPoints}
}

// Validate 設定値を検証する。
func (c AluetteConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(AluetteCpuDifficultyEasy), int(AluetteCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateMin("target points", c.TargetPoints, 1)
}
