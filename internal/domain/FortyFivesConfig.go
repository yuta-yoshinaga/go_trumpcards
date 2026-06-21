//go:build !js || !wasm || casino

package domain

// FortyFivesCpuDifficulty CPU の難易度レベル
type FortyFivesCpuDifficulty int

// Forty-Fives の CPU 難易度定数
const (
	// FortyFivesCpuDifficultyEasy 低難易度（ランダムプレイ）
	FortyFivesCpuDifficultyEasy FortyFivesCpuDifficulty = iota
	// FortyFivesCpuDifficultyNormal 中難易度（戦略プレイ）
	FortyFivesCpuDifficultyNormal
	// FortyFivesCpuDifficultyHard 高難易度（戦略プレイ）
	FortyFivesCpuDifficultyHard
)

// FortyFivesConfig オークション・フォーティファイブズのゲーム設定
type FortyFivesConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty FortyFivesCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積点。いずれかのチームがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultFortyFivesConfig デフォルト設定を返す（標準は 45 点先取）。
func DefaultFortyFivesConfig() FortyFivesConfig {
	return FortyFivesConfig{CpuDifficulty: FortyFivesCpuDifficultyNormal, TargetPoints: 45}
}

// Validate 設定値のドメインバリデーション
func (c FortyFivesConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(FortyFivesCpuDifficultyEasy), int(FortyFivesCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
