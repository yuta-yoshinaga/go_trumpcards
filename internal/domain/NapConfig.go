//go:build !js || !wasm || classic

package domain

// NapCpuDifficulty CPU の難易度レベル
type NapCpuDifficulty int

// Nap の CPU 難易度定数
const (
	// NapCpuDifficultyEasy 低難易度（ランダムプレイ）
	NapCpuDifficultyEasy NapCpuDifficulty = iota
	// NapCpuDifficultyNormal 中難易度（戦略プレイ）
	NapCpuDifficultyNormal
	// NapCpuDifficultyHard 高難易度（戦略プレイ）
	NapCpuDifficultyHard
)

// NapConfig ナップのゲーム設定
type NapConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty NapCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積チップ。いずれかのプレイヤーがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultNapConfig デフォルト設定を返す（標準は 20 チップ先取）。
func DefaultNapConfig() NapConfig {
	return NapConfig{CpuDifficulty: NapCpuDifficultyNormal, TargetPoints: 20}
}

// Validate 設定値のドメインバリデーション
func (c NapConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(NapCpuDifficultyEasy), int(NapCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
