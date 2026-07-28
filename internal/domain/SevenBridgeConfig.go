//go:build !js || !wasm || extra3

package domain

// SevenBridgeCpuDifficulty CPU の難易度レベル
type SevenBridgeCpuDifficulty int

// SevenBridge の CPU 難易度定数
const (
	// SevenBridgeCpuDifficultyEasy 低難易度
	SevenBridgeCpuDifficultyEasy SevenBridgeCpuDifficulty = iota
	// SevenBridgeCpuDifficultyNormal 中難易度
	SevenBridgeCpuDifficultyNormal
	// SevenBridgeCpuDifficultyHard 高難易度
	SevenBridgeCpuDifficultyHard
)

// SevenBridgeConfig セブンブリッジの設定
type SevenBridgeConfig struct {
	CpuDifficulty SevenBridgeCpuDifficulty `json:"cd"`
	PointLimit    int                      `json:"pl"` // ラウンドスコア累計で上限に到達したプレイヤーが勝利
}

// DefaultSevenBridgeConfig デフォルト設定を返す
func DefaultSevenBridgeConfig() SevenBridgeConfig {
	return SevenBridgeConfig{
		CpuDifficulty: SevenBridgeCpuDifficultyNormal,
		PointLimit:    100,
	}
}

// Validate 設定値のドメインバリデーション
func (c SevenBridgeConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SevenBridgeCpuDifficultyEasy), int(SevenBridgeCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
