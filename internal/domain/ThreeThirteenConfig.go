//go:build !js || !wasm || extra

package domain

// ThreeThirteenCpuDifficulty Three Thirteen の CPU 難易度レベル
type ThreeThirteenCpuDifficulty int

// Three Thirteen の CPU 難易度定数
const (
	// ThreeThirteenCpuDifficultyEasy 低難易度
	ThreeThirteenCpuDifficultyEasy ThreeThirteenCpuDifficulty = iota
	// ThreeThirteenCpuDifficultyNormal 中難易度
	ThreeThirteenCpuDifficultyNormal
	// ThreeThirteenCpuDifficultyHard 高難易度
	ThreeThirteenCpuDifficultyHard
)

// Three Thirteen のプレイヤー数の下限・上限
const (
	// ThreeThirteenMinPlayers 最小プレイヤー数
	ThreeThirteenMinPlayers = 2
	// ThreeThirteenMaxPlayers 最大プレイヤー数
	ThreeThirteenMaxPlayers = 4
	// ThreeThirteenDefaultPlayers 既定プレイヤー数（人間 1 + CPU 3）
	ThreeThirteenDefaultPlayers = 4
)

// ThreeThirteenConfig Three Thirteen（スリー・サーティーン）の設定
type ThreeThirteenConfig struct {
	// CpuDifficulty CPU の難易度
	CpuDifficulty ThreeThirteenCpuDifficulty `json:"cd"`
	// PlayerCount このゲームのプレイヤー数（2〜4）
	PlayerCount int `json:"pc"`
}

// DefaultThreeThirteenConfig デフォルト設定を返す。
// 4 人（人間 1 + CPU 3）、CPU 難易度 Normal。
func DefaultThreeThirteenConfig() ThreeThirteenConfig {
	return ThreeThirteenConfig{
		CpuDifficulty: ThreeThirteenCpuDifficultyNormal,
		PlayerCount:   ThreeThirteenDefaultPlayers,
	}
}

// Validate 設定値のドメインバリデーション
func (c ThreeThirteenConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ThreeThirteenCpuDifficultyEasy), int(ThreeThirteenCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("player count", c.PlayerCount, ThreeThirteenMinPlayers, ThreeThirteenMaxPlayers); err != nil {
		return err
	}
	return nil
}
