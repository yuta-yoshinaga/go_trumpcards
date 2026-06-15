//go:build !js || !wasm || classic

package domain

// SpoilFiveCpuDifficulty CPU の難易度レベル
type SpoilFiveCpuDifficulty int

// Spoil Five の CPU 難易度定数
const (
	// SpoilFiveCpuDifficultyEasy 低難易度（ランダムプレイ）
	SpoilFiveCpuDifficultyEasy SpoilFiveCpuDifficulty = iota
	// SpoilFiveCpuDifficultyNormal 中難易度（戦略プレイ）
	SpoilFiveCpuDifficultyNormal
	// SpoilFiveCpuDifficultyHard 高難易度（戦略プレイ）
	SpoilFiveCpuDifficultyHard
)

// SpoilFiveConfig スポイル・ファイブのゲーム設定
type SpoilFiveConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty SpoilFiveCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積得点。いずれかのプレイヤーがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultSpoilFiveConfig デフォルト設定を返す（標準は 30 点先取）。
func DefaultSpoilFiveConfig() SpoilFiveConfig {
	return SpoilFiveConfig{CpuDifficulty: SpoilFiveCpuDifficultyNormal, TargetPoints: 30}
}

// Validate 設定値のドメインバリデーション
func (c SpoilFiveConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SpoilFiveCpuDifficultyEasy), int(SpoilFiveCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
