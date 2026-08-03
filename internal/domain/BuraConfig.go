//go:build !js || !wasm || extra3

package domain

// BuraCpuDifficulty CPU の難易度レベル
type BuraCpuDifficulty int

// BuraのCPU難易度定数
const (
	// BuraCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	BuraCpuDifficultyNormal BuraCpuDifficulty = iota
)

// BuraConfig ブラのゲーム設定
type BuraConfig struct {
	CpuDifficulty BuraCpuDifficulty `json:"cd"`
}

// DefaultBuraConfig デフォルト設定を返す
func DefaultBuraConfig() BuraConfig {
	return BuraConfig{
		CpuDifficulty: BuraCpuDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c BuraConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(BuraCpuDifficultyNormal), int(BuraCpuDifficultyNormal))
}
