//go:build !js || !wasm || casino

package domain

// KlaverjasCpuDifficulty CPU の難易度レベル
type KlaverjasCpuDifficulty int

// Klaverjas の CPU 難易度定数
const (
	// KlaverjasCpuDifficultyEasy 低難易度（ランダムプレイ）
	KlaverjasCpuDifficultyEasy KlaverjasCpuDifficulty = iota
	// KlaverjasCpuDifficultyNormal 中難易度（戦略プレイ）
	KlaverjasCpuDifficultyNormal
	// KlaverjasCpuDifficultyHard 高難易度（戦略プレイ）
	KlaverjasCpuDifficultyHard
)

// KlaverjasConfig クラヴァヤスのゲーム設定
type KlaverjasConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty KlaverjasCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積点。いずれかのチームがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultKlaverjasConfig デフォルト設定を返す（標準は 1501 点先取）。
func DefaultKlaverjasConfig() KlaverjasConfig {
	return KlaverjasConfig{CpuDifficulty: KlaverjasCpuDifficultyNormal, TargetPoints: 1501}
}

// Validate 設定値のドメインバリデーション
func (c KlaverjasConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(KlaverjasCpuDifficultyEasy), int(KlaverjasCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
