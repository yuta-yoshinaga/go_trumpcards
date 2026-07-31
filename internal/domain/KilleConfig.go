//go:build !js || !wasm || extra3

package domain

// KilleCpuDifficulty CPU の難易度レベル
type KilleCpuDifficulty int

// Kille の CPU 難易度定数
const (
	// KilleCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	KilleCpuDifficultyNormal KilleCpuDifficulty = iota
)

// KilleConfig キッレのゲーム設定
type KilleConfig struct {
	CpuDifficulty KilleCpuDifficulty `json:"cd"`
	// Stake は 1 ラウンドの掛け金。
	Stake int `json:"st"`
}

// DefaultKilleConfig デフォルト設定を返す
func DefaultKilleConfig() KilleConfig {
	return KilleConfig{
		CpuDifficulty: KilleCpuDifficultyNormal,
		Stake:         1,
	}
}

// Validate 設定値のドメインバリデーション
func (c KilleConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(KilleCpuDifficultyNormal), int(KilleCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("stake", c.Stake, 1, 100)
}
