//go:build !js || !wasm || classic

package domain

// HeartsCpuDifficulty CPU の難易度レベル
type HeartsCpuDifficulty int

// HeartsのCPU難易度定数
const (
	// HeartsCpuDifficultyEasy 低難易度
	HeartsCpuDifficultyEasy HeartsCpuDifficulty = iota
	// HeartsCpuDifficultyNormal 中難易度
	HeartsCpuDifficultyNormal
	// HeartsCpuDifficultyHard 高難易度
	HeartsCpuDifficultyHard
)

// HeartsConfig ハーツゲーム設定
type HeartsConfig struct {
	CpuDifficulty HeartsCpuDifficulty `json:"cd"`
	PointLimit    int                 `json:"pl"` // ゲーム終了スコア (いずれかのプレイヤーがこの点数に達したら終了)
	OmnibusJD     bool                `json:"oj"` // オムニバス・ハーツ: J♦獲得で-10点
}

// DefaultHeartsConfig デフォルト設定を返す
func DefaultHeartsConfig() HeartsConfig {
	return HeartsConfig{CpuDifficulty: HeartsCpuDifficultyNormal, PointLimit: 100}
}

// Validate 設定値のドメインバリデーション
func (c HeartsConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(HeartsCpuDifficultyEasy), int(HeartsCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
