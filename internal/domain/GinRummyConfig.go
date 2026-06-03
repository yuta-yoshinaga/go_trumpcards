//go:build !js || !wasm || solo

package domain

// GinRummyCpuDifficulty CPU の難易度レベル
type GinRummyCpuDifficulty int

// GinRummyのCPU難易度定数
const (
	// GinRummyCpuDifficultyEasy 低難易度
	GinRummyCpuDifficultyEasy GinRummyCpuDifficulty = iota
	// GinRummyCpuDifficultyNormal 中難易度
	GinRummyCpuDifficultyNormal
	// GinRummyCpuDifficultyHard 高難易度
	GinRummyCpuDifficultyHard
)

// GinRummyConfig ジンラミーゲーム設定
type GinRummyConfig struct {
	CpuDifficulty GinRummyCpuDifficulty `json:"cd"`
	PointLimit    int                   `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultGinRummyConfig デフォルト設定を返す
func DefaultGinRummyConfig() GinRummyConfig {
	return GinRummyConfig{
		CpuDifficulty: GinRummyCpuDifficultyNormal,
		PointLimit:    100,
	}
}

// Validate 設定値のドメインバリデーション
func (c GinRummyConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(GinRummyCpuDifficultyEasy), int(GinRummyCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
