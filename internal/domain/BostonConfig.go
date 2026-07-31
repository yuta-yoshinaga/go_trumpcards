//go:build !js || !wasm || extra3

package domain

// BostonCpuDifficulty CPU の難易度レベル
type BostonCpuDifficulty int

// Boston の CPU 難易度定数
const (
	// BostonCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	BostonCpuDifficultyNormal BostonCpuDifficulty = iota
)

// BostonTargetHandsDefault は既定の規定局数。
const BostonTargetHandsDefault = 8

// BostonTargetHandsMin / BostonTargetHandsMax は規定局数の範囲。
const (
	BostonTargetHandsMin = 1
	BostonTargetHandsMax = 30
)

// BostonConfig ボストンのゲーム設定
type BostonConfig struct {
	CpuDifficulty BostonCpuDifficulty `json:"cd"`
	// TargetHands は決着までの局数。
	TargetHands int `json:"th"`
}

// DefaultBostonConfig デフォルト設定を返す
func DefaultBostonConfig() BostonConfig {
	return BostonConfig{
		CpuDifficulty: BostonCpuDifficultyNormal,
		TargetHands:   BostonTargetHandsDefault,
	}
}

// Validate 設定値のドメインバリデーション
func (c BostonConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(BostonCpuDifficultyNormal), int(BostonCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("target hands", c.TargetHands, BostonTargetHandsMin, BostonTargetHandsMax)
}
