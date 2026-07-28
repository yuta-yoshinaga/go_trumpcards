//go:build !js || !wasm || extra3

package domain

// MaoCpuDifficulty CPU の難易度レベル
type MaoCpuDifficulty int

// MaoのCPU難易度定数
const (
	// MaoCpuDifficultyEasy 低難易度
	MaoCpuDifficultyEasy MaoCpuDifficulty = iota
	// MaoCpuDifficultyNormal 中難易度
	MaoCpuDifficultyNormal
	// MaoCpuDifficultyHard 高難易度
	MaoCpuDifficultyHard
)

// MaoConfig マオゲーム設定
type MaoConfig struct {
	CpuDifficulty MaoCpuDifficulty `json:"cd"`
	PointLimit    int              `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultMaoConfig デフォルト設定を返す
func DefaultMaoConfig() MaoConfig {
	return MaoConfig{
		CpuDifficulty: MaoCpuDifficultyNormal,
		PointLimit:    200,
	}
}

// Validate 設定値のドメインバリデーション
func (c MaoConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(MaoCpuDifficultyEasy), int(MaoCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
