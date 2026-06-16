//go:build !js || !wasm || casino

package domain

// TwentyNineCpuDifficulty CPU の難易度レベル
type TwentyNineCpuDifficulty int

// 29 の CPU 難易度定数
const (
	// TwentyNineCpuDifficultyEasy 低難易度（ランダムプレイ）
	TwentyNineCpuDifficultyEasy TwentyNineCpuDifficulty = iota
	// TwentyNineCpuDifficultyNormal 中難易度（戦略プレイ）
	TwentyNineCpuDifficultyNormal
	// TwentyNineCpuDifficultyHard 高難易度（戦略プレイ）
	TwentyNineCpuDifficultyHard
)

// TwentyNineConfig トゥエンティナインのゲーム設定
type TwentyNineConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty TwentyNineCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積ゲーム点。いずれかのチームがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultTwentyNineConfig デフォルト設定を返す（標準は 6 ゲーム点先取）。
func DefaultTwentyNineConfig() TwentyNineConfig {
	return TwentyNineConfig{CpuDifficulty: TwentyNineCpuDifficultyNormal, TargetPoints: 6}
}

// Validate 設定値のドメインバリデーション
func (c TwentyNineConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TwentyNineCpuDifficultyEasy), int(TwentyNineCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
