//go:build !js || !wasm || extra5

package domain

// TongitsCpuDifficulty CPU の難易度レベル
type TongitsCpuDifficulty int

// TongitsのCPU難易度定数
const (
	// TongitsCpuDifficultyEasy 低難易度
	TongitsCpuDifficultyEasy TongitsCpuDifficulty = iota
	// TongitsCpuDifficultyNormal 中難易度
	TongitsCpuDifficultyNormal
	// TongitsCpuDifficultyHard 高難易度
	TongitsCpuDifficultyHard
)

// TongitsConfig Tongitsゲーム設定
type TongitsConfig struct {
	CpuDifficulty TongitsCpuDifficulty `json:"cd"`
	PointLimit    int                  `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultTongitsConfig デフォルト設定を返す
func DefaultTongitsConfig() TongitsConfig {
	return TongitsConfig{
		CpuDifficulty: TongitsCpuDifficultyNormal,
		PointLimit:    50,
	}
}

// Validate 設定値のドメインバリデーション
func (c TongitsConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TongitsCpuDifficultyEasy), int(TongitsCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
