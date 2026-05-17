package domain

// BriscolaCpuDifficulty CPU の難易度レベル
type BriscolaCpuDifficulty int

// BriscolaのCPU難易度定数
const (
	// BriscolaCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	BriscolaCpuDifficultyNormal BriscolaCpuDifficulty = iota
)

// BriscolaConfig ブリスコラゲーム設定
type BriscolaConfig struct {
	CpuDifficulty BriscolaCpuDifficulty `json:"cd"`
}

// DefaultBriscolaConfig デフォルト設定を返す
func DefaultBriscolaConfig() BriscolaConfig {
	return BriscolaConfig{
		CpuDifficulty: BriscolaCpuDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c BriscolaConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(BriscolaCpuDifficultyNormal), int(BriscolaCpuDifficultyNormal))
}
