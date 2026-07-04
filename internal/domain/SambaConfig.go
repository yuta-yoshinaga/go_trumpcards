//go:build !js || !wasm || extra

package domain

// SambaCpuDifficulty CPU の難易度レベル
type SambaCpuDifficulty int

// SambaのCPU難易度定数
const (
	// SambaCpuDifficultyEasy 低難易度
	SambaCpuDifficultyEasy SambaCpuDifficulty = iota
	// SambaCpuDifficultyNormal 中難易度
	SambaCpuDifficultyNormal
	// SambaCpuDifficultyHard 高難易度
	SambaCpuDifficultyHard
)

// SambaConfig サンバゲーム設定
type SambaConfig struct {
	CpuDifficulty SambaCpuDifficulty `json:"cd"`
	PointLimit    int                `json:"pl"` // ゲーム終了スコア (先に到達したチームが勝利)
}

// DefaultSambaConfig デフォルト設定を返す
func DefaultSambaConfig() SambaConfig {
	return SambaConfig{
		CpuDifficulty: SambaCpuDifficultyNormal,
		PointLimit:    SambaDefaultPointLimit,
	}
}

// Validate 設定値のドメインバリデーション
func (c SambaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SambaCpuDifficultyEasy), int(SambaCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
