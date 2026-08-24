//go:build !js || !wasm || extra4

package domain

// SchafkopfCpuDifficulty CPU の難易度レベル
type SchafkopfCpuDifficulty int

// Schafkopf の CPU 難易度定数
const (
	// SchafkopfCpuDifficultyEasy 低難易度（ランダムプレイ）
	SchafkopfCpuDifficultyEasy SchafkopfCpuDifficulty = iota
	// SchafkopfCpuDifficultyNormal 中難易度（戦略プレイ）
	SchafkopfCpuDifficultyNormal
	// SchafkopfCpuDifficultyHard 高難易度（戦略プレイ）
	SchafkopfCpuDifficultyHard
)

// SchafkopfConfig シャーフコップのゲーム設定
type SchafkopfConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty SchafkopfCpuDifficulty `json:"cd"`
	// BaseChips 1ラウンドの基本チップ単位。各ラウンドの精算はこの値の倍数で行われる。
	BaseChips int `json:"bc"`
	// StartChips 各プレイヤーの開始チップ数。
	StartChips int `json:"sc"`
	// TargetChips ゲーム終了に必要なチップ数。いずれかのプレイヤーの所持チップが
	// この値以上に達したらゲーム終了。
	TargetChips int `json:"tc"`
}

// DefaultSchafkopfConfig デフォルト設定を返す。
func DefaultSchafkopfConfig() SchafkopfConfig {
	return SchafkopfConfig{
		CpuDifficulty: SchafkopfCpuDifficultyNormal,
		BaseChips:     2,
		StartChips:    20,
		TargetChips:   40,
	}
}

// Validate 設定値のドメインバリデーション
func (c SchafkopfConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SchafkopfCpuDifficultyEasy), int(SchafkopfCpuDifficultyHard)); err != nil {
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
