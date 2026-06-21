//go:build !js || !wasm || classic

package domain

// MariasCpuDifficulty CPU の難易度レベル
type MariasCpuDifficulty int

// Mariáš の CPU 難易度定数
const (
	// MariasCpuDifficultyEasy 低難易度（ランダムプレイ）
	MariasCpuDifficultyEasy MariasCpuDifficulty = iota
	// MariasCpuDifficultyNormal 中難易度（戦略プレイ）
	MariasCpuDifficultyNormal
	// MariasCpuDifficultyHard 高難易度（戦略プレイ）
	MariasCpuDifficultyHard
)

// MariasConfig マリアーシュのゲーム設定
type MariasConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty MariasCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積ゲーム点。いずれかのプレイヤーがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultMariasConfig デフォルト設定を返す（標準は 10 点先取）。
func DefaultMariasConfig() MariasConfig {
	return MariasConfig{CpuDifficulty: MariasCpuDifficultyNormal, TargetPoints: 10}
}

// Validate 設定値のドメインバリデーション
func (c MariasConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(MariasCpuDifficultyEasy), int(MariasCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
