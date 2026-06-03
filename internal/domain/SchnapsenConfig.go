package domain

// SchnapsenCpuDifficulty CPU の難易度レベル
type SchnapsenCpuDifficulty int

// SchnapsenのCPU難易度定数
const (
	// SchnapsenCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	SchnapsenCpuDifficultyNormal SchnapsenCpuDifficulty = iota
)

// SchnapsenConfig シュナプセン (Schnapsen / Sixty-Six) ゲーム設定
type SchnapsenConfig struct {
	CpuDifficulty SchnapsenCpuDifficulty `json:"cd"`
}

// DefaultSchnapsenConfig デフォルト設定を返す
func DefaultSchnapsenConfig() SchnapsenConfig {
	return SchnapsenConfig{
		CpuDifficulty: SchnapsenCpuDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c SchnapsenConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(SchnapsenCpuDifficultyNormal), int(SchnapsenCpuDifficultyNormal))
}
