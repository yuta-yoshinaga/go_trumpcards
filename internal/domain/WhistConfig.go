package domain

// WhistCpuDifficulty CPU の難易度レベル
type WhistCpuDifficulty int

// WhistのCPU難易度定数
const (
	// WhistCpuDifficultyEasy 低難易度
	WhistCpuDifficultyEasy WhistCpuDifficulty = iota
	// WhistCpuDifficultyNormal 中難易度
	WhistCpuDifficultyNormal
	// WhistCpuDifficultyHard 高難易度
	WhistCpuDifficultyHard
)

// WhistConfig ホイストゲーム設定
type WhistConfig struct {
	CpuDifficulty WhistCpuDifficulty `json:"cd"`
	PointLimit    int                `json:"pl"` // ゲーム終了スコア (先に到達したチームが勝利, デフォルト5)
}

// DefaultWhistConfig デフォルト設定を返す
func DefaultWhistConfig() WhistConfig {
	return WhistConfig{
		CpuDifficulty: WhistCpuDifficultyNormal,
		PointLimit:    5,
	}
}

// Validate 設定値のドメインバリデーション
func (c WhistConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(WhistCpuDifficultyEasy), int(WhistCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
