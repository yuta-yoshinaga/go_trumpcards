package domain

import "fmt"

// NapoleonCpuDifficulty CPU の難易度レベル
type NapoleonCpuDifficulty int

// ナポレオンのCPU難易度定数
const (
	// NapoleonCpuDifficultyEasy 低難易度
	NapoleonCpuDifficultyEasy NapoleonCpuDifficulty = iota
	// NapoleonCpuDifficultyNormal 中難易度
	NapoleonCpuDifficultyNormal
	// NapoleonCpuDifficultyHard 高難易度
	NapoleonCpuDifficultyHard
)

// NapoleonConfig ナポレオンゲーム設定
type NapoleonConfig struct {
	CpuDifficulty NapoleonCpuDifficulty
	MinBid        int // 最低ビッド値 (デフォルト12)
	PointLimit    int // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultNapoleonConfig デフォルト設定を返す
func DefaultNapoleonConfig() NapoleonConfig {
	return NapoleonConfig{
		CpuDifficulty: NapoleonCpuDifficultyNormal,
		MinBid:        12,
		PointLimit:    100,
	}
}

// Validate 設定値のドメインバリデーション
func (c NapoleonConfig) Validate() error {
	if c.CpuDifficulty < NapoleonCpuDifficultyEasy || c.CpuDifficulty > NapoleonCpuDifficultyHard {
		return fmt.Errorf("CPU difficulty must be %d-%d, got %d", int(NapoleonCpuDifficultyEasy), int(NapoleonCpuDifficultyHard), int(c.CpuDifficulty))
	}
	if c.MinBid < 1 || c.MinBid > NapoleonMaxPictureCards {
		return fmt.Errorf("min bid must be 1-%d, got %d", NapoleonMaxPictureCards, c.MinBid)
	}
	if c.PointLimit < 1 {
		return fmt.Errorf("point limit must be >= 1, got %d", c.PointLimit)
	}
	return nil
}
