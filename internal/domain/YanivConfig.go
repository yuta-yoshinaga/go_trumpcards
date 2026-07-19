//go:build !js || !wasm || solo

package domain

// YanivCpuDifficulty CPU の難易度レベル
type YanivCpuDifficulty int

// Yaniv の CPU 難易度定数
const (
	// YanivCpuDifficultyEasy 低難易度
	YanivCpuDifficultyEasy YanivCpuDifficulty = iota
	// YanivCpuDifficultyNormal 中難易度
	YanivCpuDifficultyNormal
	// YanivCpuDifficultyHard 高難易度
	YanivCpuDifficultyHard
)

// YanivMinScoreLimit 設定可能な脱落スコアの下限
const YanivMinScoreLimit = 50

// YanivMaxScoreLimit 設定可能な脱落スコアの上限
const YanivMaxScoreLimit = 500

// YanivConfig Yaniv ゲーム設定
type YanivConfig struct {
	CpuDifficulty YanivCpuDifficulty `json:"cd"`
	ScoreLimit    int                `json:"sl"` // この累計点を超えたプレイヤーは脱落
}

// DefaultYanivConfig デフォルト設定を返す
func DefaultYanivConfig() YanivConfig {
	return YanivConfig{
		CpuDifficulty: YanivCpuDifficultyNormal,
		ScoreLimit:    200,
	}
}

// Validate 設定値のドメインバリデーション
func (c YanivConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(YanivCpuDifficultyEasy), int(YanivCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("score limit", c.ScoreLimit, YanivMinScoreLimit, YanivMaxScoreLimit); err != nil {
		return err
	}
	return nil
}
