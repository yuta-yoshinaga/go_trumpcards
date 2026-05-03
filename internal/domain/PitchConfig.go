package domain

// PitchCpuDifficulty CPU の難易度レベル
type PitchCpuDifficulty int

// PitchのCPU難易度定数
const (
	// PitchCpuDifficultyEasy 低難易度
	PitchCpuDifficultyEasy PitchCpuDifficulty = iota
	// PitchCpuDifficultyNormal 中難易度
	PitchCpuDifficultyNormal
	// PitchCpuDifficultyHard 高難易度
	PitchCpuDifficultyHard
)

// PitchConfig ピッチ (Auction Pitch / Setback) ゲーム設定
type PitchConfig struct {
	CpuDifficulty PitchCpuDifficulty `json:"cd"`
	PointLimit    int                `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利, デフォルト7)
}

// DefaultPitchConfig デフォルト設定を返す
func DefaultPitchConfig() PitchConfig {
	return PitchConfig{
		CpuDifficulty: PitchCpuDifficultyNormal,
		PointLimit:    7,
	}
}

// Validate 設定値のドメインバリデーション
func (c PitchConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(PitchCpuDifficultyEasy), int(PitchCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
