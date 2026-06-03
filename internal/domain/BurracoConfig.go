package domain

// BurracoCpuDifficulty CPU の難易度レベル
type BurracoCpuDifficulty int

// BurracoのCPU難易度定数
const (
	// BurracoCpuDifficultyEasy 低難易度
	BurracoCpuDifficultyEasy BurracoCpuDifficulty = iota
	// BurracoCpuDifficultyNormal 中難易度
	BurracoCpuDifficultyNormal
	// BurracoCpuDifficultyHard 高難易度
	BurracoCpuDifficultyHard
)

// BurracoConfig ブラーコゲーム設定
type BurracoConfig struct {
	CpuDifficulty BurracoCpuDifficulty `json:"cd"`
	PointLimit    int                  `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultBurracoConfig デフォルト設定を返す
func DefaultBurracoConfig() BurracoConfig {
	return BurracoConfig{
		CpuDifficulty: BurracoCpuDifficultyNormal,
		PointLimit:    BurracoDefaultPointLimit,
	}
}

// Validate 設定値のドメインバリデーション
func (c BurracoConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(BurracoCpuDifficultyEasy), int(BurracoCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
