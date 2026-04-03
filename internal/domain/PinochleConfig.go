package domain

// PinochleCpuDifficulty CPU の難易度レベル
type PinochleCpuDifficulty int

// ピノクルのCPU難易度定数
const (
	// PinochleCpuDifficultyEasy 低難易度
	PinochleCpuDifficultyEasy PinochleCpuDifficulty = iota
	// PinochleCpuDifficultyNormal 中難易度
	PinochleCpuDifficultyNormal
	// PinochleCpuDifficultyHard 高難易度
	PinochleCpuDifficultyHard
)

// PinochleConfig ピノクルゲーム設定
type PinochleConfig struct {
	CpuDifficulty PinochleCpuDifficulty `json:"cd"`
	PointLimit    int                   `json:"pl"` // ゲーム終了スコア (先に到達したチームが勝利, デフォルト1500)
}

// DefaultPinochleConfig デフォルト設定を返す
func DefaultPinochleConfig() PinochleConfig {
	return PinochleConfig{
		CpuDifficulty: PinochleCpuDifficultyNormal,
		PointLimit:    1500,
	}
}

// Validate 設定値のドメインバリデーション
func (c PinochleConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(PinochleCpuDifficultyEasy), int(PinochleCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
