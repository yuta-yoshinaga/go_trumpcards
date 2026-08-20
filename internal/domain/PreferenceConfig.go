//go:build !js || !wasm || extra4

package domain

// PreferenceCpuDifficulty CPU の難易度レベル
type PreferenceCpuDifficulty int

// Préférence の CPU 難易度定数
const (
	// PreferenceCpuDifficultyEasy 低難易度（ランダムプレイ）
	PreferenceCpuDifficultyEasy PreferenceCpuDifficulty = iota
	// PreferenceCpuDifficultyNormal 中難易度（戦略プレイ）
	PreferenceCpuDifficultyNormal
	// PreferenceCpuDifficultyHard 高難易度（戦略プレイ）
	PreferenceCpuDifficultyHard
)

// PreferenceConfig プレフェランスのゲーム設定
type PreferenceConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty PreferenceCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積ゲーム点。いずれかのプレイヤーがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultPreferenceConfig デフォルト設定を返す（標準は 30 点先取）。
func DefaultPreferenceConfig() PreferenceConfig {
	return PreferenceConfig{CpuDifficulty: PreferenceCpuDifficultyNormal, TargetPoints: 30}
}

// Validate 設定値のドメインバリデーション
func (c PreferenceConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(PreferenceCpuDifficultyEasy), int(PreferenceCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
