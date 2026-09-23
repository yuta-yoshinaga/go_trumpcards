package domain

// BinokelCpuDifficulty CPU の難易度レベル
type BinokelCpuDifficulty int

// ビノクルのCPU難易度定数
const (
	// BinokelCpuDifficultyEasy 低難易度
	BinokelCpuDifficultyEasy BinokelCpuDifficulty = iota
	// BinokelCpuDifficultyNormal 中難易度
	BinokelCpuDifficultyNormal
	// BinokelCpuDifficultyHard 高難易度
	BinokelCpuDifficultyHard
)

// BinokelConfig ビノクルゲーム設定
type BinokelConfig struct {
	CpuDifficulty BinokelCpuDifficulty `json:"cd"`
	PointLimit    int                  `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利, デフォルト1000)
}

// DefaultBinokelConfig デフォルト設定を返す
func DefaultBinokelConfig() BinokelConfig {
	return BinokelConfig{
		CpuDifficulty: BinokelCpuDifficultyNormal,
		PointLimit:    1000,
	}
}

// Validate 設定値のドメインバリデーション
func (c BinokelConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(BinokelCpuDifficultyEasy), int(BinokelCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
