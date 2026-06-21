//go:build !js || !wasm || classic

package domain

// KnockoutWhistCpuDifficulty CPU の難易度レベル
type KnockoutWhistCpuDifficulty int

// Knockout Whist の CPU 難易度定数
const (
	// KnockoutWhistCpuDifficultyEasy 低難易度（ランダムプレイ）
	KnockoutWhistCpuDifficultyEasy KnockoutWhistCpuDifficulty = iota
	// KnockoutWhistCpuDifficultyNormal 中難易度（戦略プレイ）
	KnockoutWhistCpuDifficultyNormal
	// KnockoutWhistCpuDifficultyHard 高難易度（戦略プレイ）
	KnockoutWhistCpuDifficultyHard
)

// KnockoutWhistConfig ノックアウト・ホイストのゲーム設定
type KnockoutWhistConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty KnockoutWhistCpuDifficulty `json:"cd"`
}

// DefaultKnockoutWhistConfig デフォルト設定を返す。
func DefaultKnockoutWhistConfig() KnockoutWhistConfig {
	return KnockoutWhistConfig{CpuDifficulty: KnockoutWhistCpuDifficultyNormal}
}

// Validate 設定値のドメインバリデーション
func (c KnockoutWhistConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(KnockoutWhistCpuDifficultyEasy), int(KnockoutWhistCpuDifficultyHard))
}
