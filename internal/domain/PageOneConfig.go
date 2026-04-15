package domain

// PageOneCpuDifficulty CPU の難易度レベル
type PageOneCpuDifficulty int

// PageOneのCPU難易度定数
const (
	// PageOneCpuDifficultyEasy 低難易度
	PageOneCpuDifficultyEasy PageOneCpuDifficulty = iota
	// PageOneCpuDifficultyNormal 中難易度
	PageOneCpuDifficultyNormal
	// PageOneCpuDifficultyHard 高難易度
	PageOneCpuDifficultyHard
)

// PageOneConfig ページワンゲーム設定
type PageOneConfig struct {
	CpuDifficulty PageOneCpuDifficulty `json:"cd"`
	PointLimit    int                  `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultPageOneConfig デフォルト設定を返す
func DefaultPageOneConfig() PageOneConfig {
	return PageOneConfig{
		CpuDifficulty: PageOneCpuDifficultyNormal,
		PointLimit:    200,
	}
}

// Validate 設定値のドメインバリデーション
func (c PageOneConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(PageOneCpuDifficultyEasy), int(PageOneCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
