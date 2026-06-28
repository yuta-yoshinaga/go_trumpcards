package domain

// PrsiCpuDifficulty CPU の難易度レベル
type PrsiCpuDifficulty int

// PrsiのCPU難易度定数
const (
	// PrsiCpuDifficultyEasy 低難易度
	PrsiCpuDifficultyEasy PrsiCpuDifficulty = iota
	// PrsiCpuDifficultyNormal 中難易度
	PrsiCpuDifficultyNormal
	// PrsiCpuDifficultyHard 高難易度
	PrsiCpuDifficultyHard
)

// PrsiConfig プルシー(チェコ版クレイジーエイト)ゲーム設定
type PrsiConfig struct {
	CpuDifficulty PrsiCpuDifficulty `json:"cd"`
}

// DefaultPrsiConfig デフォルト設定を返す
func DefaultPrsiConfig() PrsiConfig {
	return PrsiConfig{
		CpuDifficulty: PrsiCpuDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c PrsiConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(PrsiCpuDifficultyEasy), int(PrsiCpuDifficultyHard))
}
