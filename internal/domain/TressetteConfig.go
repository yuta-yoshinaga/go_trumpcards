//go:build !js || !wasm || casino

package domain

// TressetteCpuDifficulty CPU の難易度レベル
type TressetteCpuDifficulty int

// Tressette の CPU 難易度定数
const (
	// TressetteCpuDifficultyEasy 低難易度（ランダムプレイ）
	TressetteCpuDifficultyEasy TressetteCpuDifficulty = iota
	// TressetteCpuDifficultyNormal 中難易度（戦略プレイ）
	TressetteCpuDifficultyNormal
	// TressetteCpuDifficultyHard 高難易度（戦略プレイ）
	TressetteCpuDifficultyHard
)

// TressetteConfig トレセッテのゲーム設定
type TressetteConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty TressetteCpuDifficulty `json:"cd"`
	// TargetPoints ゲーム勝利に必要な得点。いずれかのチームの累積得点が
	// この値以上に達し、かつ相手チームを上回ったらゲーム終了。
	TargetPoints int `json:"tp"`
}

// DefaultTressetteConfig デフォルト設定を返す（標準は 21 点先取）。
func DefaultTressetteConfig() TressetteConfig {
	return TressetteConfig{CpuDifficulty: TressetteCpuDifficultyNormal, TargetPoints: 21}
}

// Validate 設定値のドメインバリデーション
func (c TressetteConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TressetteCpuDifficultyEasy), int(TressetteCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
