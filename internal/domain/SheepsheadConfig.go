//go:build !js || !wasm || extra3

package domain

// SheepsheadCpuDifficulty CPU の難易度レベル
type SheepsheadCpuDifficulty int

// Sheepshead の CPU 難易度定数
const (
	// SheepsheadCpuDifficultyEasy 低難易度（ランダムプレイ）
	SheepsheadCpuDifficultyEasy SheepsheadCpuDifficulty = iota
	// SheepsheadCpuDifficultyNormal 中難易度（戦略プレイ）
	SheepsheadCpuDifficultyNormal
	// SheepsheadCpuDifficultyHard 高難易度（戦略プレイ）
	SheepsheadCpuDifficultyHard
)

// SheepsheadConfig シープスヘッドのゲーム設定
type SheepsheadConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty SheepsheadCpuDifficulty `json:"cd"`
	// BaseChips 1ラウンドの基本チップ単位。各ラウンドの精算はこの値の倍数で行われる。
	BaseChips int `json:"bc"`
	// StartChips 各プレイヤーの開始チップ数。
	StartChips int `json:"sc"`
	// TargetChips ゲーム終了に必要なチップ数。いずれかのプレイヤーの所持チップが
	// この値以上に達したらゲーム終了。
	TargetChips int `json:"tc"`
}

// DefaultSheepsheadConfig デフォルト設定を返す。
func DefaultSheepsheadConfig() SheepsheadConfig {
	return SheepsheadConfig{
		CpuDifficulty: SheepsheadCpuDifficultyNormal,
		BaseChips:     2,
		StartChips:    20,
		TargetChips:   40,
	}
}

// Validate 設定値のドメインバリデーション
func (c SheepsheadConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SheepsheadCpuDifficultyEasy), int(SheepsheadCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("base chips", c.BaseChips, 1); err != nil {
		return err
	}
	if err := ValidateMin("start chips", c.StartChips, 1); err != nil {
		return err
	}
	if err := ValidateMin("target chips", c.TargetChips, c.StartChips+1); err != nil {
		return err
	}
	return nil
}
