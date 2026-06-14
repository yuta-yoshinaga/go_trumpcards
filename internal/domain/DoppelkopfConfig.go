//go:build !js || !wasm || casino

package domain

// DoppelkopfCpuDifficulty CPU の難易度レベル
type DoppelkopfCpuDifficulty int

// Doppelkopf の CPU 難易度定数
const (
	// DoppelkopfCpuDifficultyEasy 低難易度（ランダムプレイ）
	DoppelkopfCpuDifficultyEasy DoppelkopfCpuDifficulty = iota
	// DoppelkopfCpuDifficultyNormal 中難易度（戦略プレイ）
	DoppelkopfCpuDifficultyNormal
	// DoppelkopfCpuDifficultyHard 高難易度（戦略プレイ）
	DoppelkopfCpuDifficultyHard
)

// DoppelkopfConfig ドッペルコップのゲーム設定
type DoppelkopfConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty DoppelkopfCpuDifficulty `json:"cd"`
	// BaseChips 1ゲームポイントあたりのチップ単位。
	BaseChips int `json:"bc"`
	// StartChips 各プレイヤーの開始チップ数。
	StartChips int `json:"sc"`
	// TargetChips ゲーム終了に必要なチップ数。いずれかのプレイヤーの所持チップが
	// この値以上に達したらゲーム終了。
	TargetChips int `json:"tc"`
}

// DefaultDoppelkopfConfig デフォルト設定を返す。
func DefaultDoppelkopfConfig() DoppelkopfConfig {
	return DoppelkopfConfig{
		CpuDifficulty: DoppelkopfCpuDifficultyNormal,
		BaseChips:     2,
		StartChips:    20,
		TargetChips:   40,
	}
}

// Validate 設定値のドメインバリデーション
func (c DoppelkopfConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(DoppelkopfCpuDifficultyEasy), int(DoppelkopfCpuDifficultyHard)); err != nil {
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
