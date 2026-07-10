package domain

// WizardCpuDifficulty CPU の難易度レベル
type WizardCpuDifficulty int

// WizardのCPU難易度定数
const (
	// WizardCpuDifficultyEasy 低難易度
	WizardCpuDifficultyEasy WizardCpuDifficulty = iota
	// WizardCpuDifficultyNormal 中難易度
	WizardCpuDifficultyNormal
	// WizardCpuDifficultyHard 高難易度
	WizardCpuDifficultyHard
)

// WizardConfig ウィザードゲーム設定。
// ラウンド数(15)・手札枚数(=ラウンド番号)・スコアリング方式は固定のため、
// 設定項目はCPU難易度のみ。
type WizardConfig struct {
	CpuDifficulty WizardCpuDifficulty `json:"cd"`
}

// DefaultWizardConfig デフォルト設定を返す
func DefaultWizardConfig() WizardConfig {
	return WizardConfig{
		CpuDifficulty: WizardCpuDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c WizardConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(WizardCpuDifficultyEasy), int(WizardCpuDifficultyHard)); err != nil {
		return err
	}
	return nil
}
