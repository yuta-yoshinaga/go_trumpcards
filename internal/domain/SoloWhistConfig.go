//go:build !js || !wasm || classic

package domain

// SoloWhistCpuDifficulty CPU の難易度レベル
type SoloWhistCpuDifficulty int

// Solo Whist の CPU 難易度定数
const (
	// SoloWhistCpuDifficultyEasy 低難易度（ランダムプレイ）
	SoloWhistCpuDifficultyEasy SoloWhistCpuDifficulty = iota
	// SoloWhistCpuDifficultyNormal 中難易度（戦略プレイ）
	SoloWhistCpuDifficultyNormal
	// SoloWhistCpuDifficultyHard 高難易度（戦略プレイ）
	SoloWhistCpuDifficultyHard
)

// SoloWhistConfig ソロ・ホイストのゲーム設定
type SoloWhistConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty SoloWhistCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積ゲーム点。いずれかのプレイヤーがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultSoloWhistConfig デフォルト設定を返す（標準は 21 点先取）。
func DefaultSoloWhistConfig() SoloWhistConfig {
	return SoloWhistConfig{CpuDifficulty: SoloWhistCpuDifficultyNormal, TargetPoints: 21}
}

// Validate 設定値のドメインバリデーション
func (c SoloWhistConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SoloWhistCpuDifficultyEasy), int(SoloWhistCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
