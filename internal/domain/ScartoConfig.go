//go:build !js || !wasm || extra4

package domain

// ScartoCpuDifficulty CPU の難易度レベル
type ScartoCpuDifficulty int

// Scarto の CPU 難易度定数
const (
	// ScartoCpuDifficultyEasy 低難易度 (ランダムプレイ)
	ScartoCpuDifficultyEasy ScartoCpuDifficulty = iota
	// ScartoCpuDifficultyNormal 中難易度 (戦略プレイ)
	ScartoCpuDifficultyNormal
	// ScartoCpuDifficultyHard 高難易度 (戦略プレイ)
	ScartoCpuDifficultyHard
)

// ScartoConfig スカルト (Scarto) のゲーム設定
type ScartoConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty ScartoCpuDifficulty `json:"cd"`
	// TargetDeals マッチを構成するディール数。この回数だけ配り、累積得点最上位が勝者。
	TargetDeals int `json:"td"`
}

// DefaultScartoConfig デフォルト設定を返す (標準は 3 ディール)。
func DefaultScartoConfig() ScartoConfig {
	return ScartoConfig{CpuDifficulty: ScartoCpuDifficultyNormal, TargetDeals: ScartoDefaultDeals}
}

// Validate 設定値のドメインバリデーション
func (c ScartoConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ScartoCpuDifficultyEasy), int(ScartoCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target deals", c.TargetDeals, 1); err != nil {
		return err
	}
	return nil
}
