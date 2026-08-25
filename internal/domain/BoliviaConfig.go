//go:build !js || !wasm || extra

package domain

// BoliviaCpuDifficulty CPU の難易度レベル
type BoliviaCpuDifficulty int

// BoliviaのCPU難易度定数
const (
	// BoliviaCpuDifficultyEasy 低難易度
	BoliviaCpuDifficultyEasy BoliviaCpuDifficulty = iota
	// BoliviaCpuDifficultyNormal 中難易度
	BoliviaCpuDifficultyNormal
	// BoliviaCpuDifficultyHard 高難易度
	BoliviaCpuDifficultyHard
)

// BoliviaConfig ボリビアゲーム設定
type BoliviaConfig struct {
	CpuDifficulty BoliviaCpuDifficulty `json:"cd"`
	PointLimit    int                  `json:"pl"` // ゲーム終了スコア (先に到達したチームが勝利)
}

// DefaultBoliviaConfig デフォルト設定を返す
func DefaultBoliviaConfig() BoliviaConfig {
	return BoliviaConfig{
		CpuDifficulty: BoliviaCpuDifficultyNormal,
		PointLimit:    BoliviaDefaultPointLimit,
	}
}

// Validate 設定値のドメインバリデーション
func (c BoliviaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(BoliviaCpuDifficultyEasy), int(BoliviaCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
