//go:build !js || !wasm || extra2

package domain

// ZwickerCpuDifficulty CPU の難易度レベル
type ZwickerCpuDifficulty int

// Zwicker の CPU 難易度定数
const (
	// ZwickerCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	ZwickerCpuDifficultyNormal ZwickerCpuDifficulty = iota
)

// ZwickerConfig ツヴィッカーのゲーム設定
type ZwickerConfig struct {
	CpuDifficulty ZwickerCpuDifficulty `json:"cd"`
	// TargetScore これに達したチームが勝ち。
	TargetScore int `json:"ts"`
}

// DefaultZwickerConfig デフォルト設定を返す
func DefaultZwickerConfig() ZwickerConfig {
	return ZwickerConfig{
		CpuDifficulty: ZwickerCpuDifficultyNormal,
		TargetScore:   61,
	}
}

// Validate 設定値のドメインバリデーション
func (c ZwickerConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(ZwickerCpuDifficultyNormal), int(ZwickerCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, 1, 1000)
}
