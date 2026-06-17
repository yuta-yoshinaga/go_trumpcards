//go:build !js || !wasm || classic

package domain

// ScoponeCpuDifficulty CPU の難易度レベル
type ScoponeCpuDifficulty int

// Scopone の CPU 難易度定数
const (
	// ScoponeCpuDifficultyEasy 低難易度
	ScoponeCpuDifficultyEasy ScoponeCpuDifficulty = iota
	// ScoponeCpuDifficultyNormal 中難易度
	ScoponeCpuDifficultyNormal
	// ScoponeCpuDifficultyHard 高難易度
	ScoponeCpuDifficultyHard
)

// ScoponeDefaultTargetScore 試合終了スコア (先に到達したチームが勝利)
const ScoponeDefaultTargetScore = 11

// ScoponeMaxTargetScore Validate で許容する TargetScore の上限
const ScoponeMaxTargetScore = 100

// ScoponeConfig Scopone ゲーム設定
type ScoponeConfig struct {
	CpuDifficulty ScoponeCpuDifficulty `json:"cd"`
	TargetScore   int                  `json:"ts"` // 試合終了スコア (デフォルト 11)
}

// DefaultScoponeConfig デフォルト設定を返す
func DefaultScoponeConfig() ScoponeConfig {
	return ScoponeConfig{
		CpuDifficulty: ScoponeCpuDifficultyNormal,
		TargetScore:   ScoponeDefaultTargetScore,
	}
}

// Validate 設定値のドメインバリデーション
func (c ScoponeConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ScoponeCpuDifficultyEasy), int(ScoponeCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("target score", c.TargetScore, 1, ScoponeMaxTargetScore); err != nil {
		return err
	}
	return nil
}
