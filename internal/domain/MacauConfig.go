//go:build !js || !wasm || solo

package domain

// MacauCpuDifficulty CPU の難易度レベル
type MacauCpuDifficulty int

// MacauのCPU難易度定数
const (
	// MacauCpuDifficultyEasy 低難易度
	MacauCpuDifficultyEasy MacauCpuDifficulty = iota
	// MacauCpuDifficultyNormal 中難易度
	MacauCpuDifficultyNormal
	// MacauCpuDifficultyHard 高難易度
	MacauCpuDifficultyHard
)

// MacauConfig マカオゲーム設定
type MacauConfig struct {
	CpuDifficulty MacauCpuDifficulty `json:"cd"`
	PointLimit    int                `json:"pl"` // ゲーム終了スコア (先に到達したプレイヤーが勝利)
}

// DefaultMacauConfig デフォルト設定を返す
func DefaultMacauConfig() MacauConfig {
	return MacauConfig{
		CpuDifficulty: MacauCpuDifficultyNormal,
		PointLimit:    200,
	}
}

// Validate 設定値のドメインバリデーション
func (c MacauConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(MacauCpuDifficultyEasy), int(MacauCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	return nil
}
