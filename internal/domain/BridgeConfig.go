//go:build !js || !wasm || extra3

package domain

// BridgeCpuDifficulty CPU の難易度レベル
type BridgeCpuDifficulty int

// ブリッジのCPU難易度定数
const (
	// BridgeCpuDifficultyEasy 低難易度 (ランダムビッド・プレイ)
	BridgeCpuDifficultyEasy BridgeCpuDifficulty = iota
	// BridgeCpuDifficultyNormal 中難易度 (ポイントカウント制ビッド)
	BridgeCpuDifficultyNormal
	// BridgeCpuDifficultyHard 高難易度 (戦略的プレイ)
	BridgeCpuDifficultyHard
)

// BridgeConfig ブリッジゲーム設定
type BridgeConfig struct {
	CpuDifficulty BridgeCpuDifficulty `json:"cd"`
}

// DefaultBridgeConfig デフォルト設定を返す
func DefaultBridgeConfig() BridgeConfig {
	return BridgeConfig{
		CpuDifficulty: BridgeCpuDifficultyNormal,
	}
}

// Validate 設定値のドメインバリデーション
func (c BridgeConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(BridgeCpuDifficultyEasy), int(BridgeCpuDifficultyHard)); err != nil {
		return err
	}
	return nil
}
