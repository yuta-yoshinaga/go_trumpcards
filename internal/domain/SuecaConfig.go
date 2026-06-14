//go:build !js || !wasm || casino

package domain

// SuecaCpuDifficulty CPU の難易度レベル
type SuecaCpuDifficulty int

// Sueca の CPU 難易度定数
const (
	// SuecaCpuDifficultyEasy 低難易度（ランダムプレイ）
	SuecaCpuDifficultyEasy SuecaCpuDifficulty = iota
	// SuecaCpuDifficultyNormal 中難易度（戦略プレイ）
	SuecaCpuDifficultyNormal
	// SuecaCpuDifficultyHard 高難易度（戦略プレイ）
	SuecaCpuDifficultyHard
)

// SuecaConfig スエカのゲーム設定
type SuecaConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty SuecaCpuDifficulty `json:"cd"`
	// TargetGamePoints マッチ勝利に必要なゲームポイント (jogos)。各ラウンド勝者が
	// 1/2/4 ポイントを獲得し、この値以上に達したチームが勝利。
	TargetGamePoints int `json:"tg"`
}

// DefaultSuecaConfig デフォルト設定を返す（標準は 4 ゲームポイント先取）。
func DefaultSuecaConfig() SuecaConfig {
	return SuecaConfig{CpuDifficulty: SuecaCpuDifficultyNormal, TargetGamePoints: 4}
}

// Validate 設定値のドメインバリデーション
func (c SuecaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SuecaCpuDifficultyEasy), int(SuecaCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target game points", c.TargetGamePoints, 1); err != nil {
		return err
	}
	return nil
}
