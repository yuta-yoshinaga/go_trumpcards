//go:build !js || !wasm || extra

package domain

// Rummy500CpuDifficulty CPUの難易度レベル
type Rummy500CpuDifficulty int

// Rummy500のCPU難易度定数
const (
	// Rummy500CpuDifficultyEasy 低難易度
	Rummy500CpuDifficultyEasy Rummy500CpuDifficulty = iota
	// Rummy500CpuDifficultyNormal 中難易度
	Rummy500CpuDifficultyNormal
	// Rummy500CpuDifficultyHard 高難易度
	Rummy500CpuDifficultyHard
)

// Rummy500Config Rummy 500ゲーム設定
type Rummy500Config struct {
	CpuDifficulty Rummy500CpuDifficulty `json:"cd"`
	PointLimit    int                   `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultRummy500Config デフォルト設定を返す
func DefaultRummy500Config() Rummy500Config {
	return Rummy500Config{
		CpuDifficulty: Rummy500CpuDifficultyNormal,
		PointLimit:    500,
	}
}

// Validate 設定値のドメインバリデーション
func (c Rummy500Config) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(Rummy500CpuDifficultyEasy), int(Rummy500CpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
