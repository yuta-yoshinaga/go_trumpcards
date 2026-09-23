//go:build !js || !wasm || extra5

package domain

// OmiCpuDifficulty CPU の難易度レベル
type OmiCpuDifficulty int

// OmiのCPU難易度定数
const (
	// OmiCpuDifficultyEasy 低難易度
	OmiCpuDifficultyEasy OmiCpuDifficulty = iota
	// OmiCpuDifficultyNormal 中難易度
	OmiCpuDifficultyNormal
	// OmiCpuDifficultyHard 高難易度
	OmiCpuDifficultyHard
)

// OmiConfig オミゲーム設定
type OmiConfig struct {
	CpuDifficulty OmiCpuDifficulty `json:"cd"`
	PointLimit    int              `json:"pl"` // ゲーム終了スコア (先に到達したチームが勝利, デフォルト10)
}

// DefaultOmiConfig デフォルト設定を返す
func DefaultOmiConfig() OmiConfig {
	return OmiConfig{
		CpuDifficulty: OmiCpuDifficultyNormal,
		PointLimit:    10,
	}
}

// Validate 設定値のドメインバリデーション
func (c OmiConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(OmiCpuDifficultyEasy), int(OmiCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
