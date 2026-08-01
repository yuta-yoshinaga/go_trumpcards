//go:build !js || !wasm || classic

package domain

// KarnoffelCpuDifficulty CPU の難易度レベル
type KarnoffelCpuDifficulty int

// カルニッフェルの CPU 難易度定数
const (
	// KarnoffelCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	KarnoffelCpuDifficultyNormal KarnoffelCpuDifficulty = iota
)

// 勝利に要る局数の範囲。
const (
	// KarnoffelMinTarget 最少
	KarnoffelMinTarget = 1
	// KarnoffelMaxTarget 最多
	KarnoffelMaxTarget = 10
)

// KarnoffelConfig カルニッフェルのゲーム設定
type KarnoffelConfig struct {
	CpuDifficulty KarnoffelCpuDifficulty `json:"cd"`
	// TargetHands は勝利に要る局数。
	TargetHands int `json:"th"`
}

// DefaultKarnoffelConfig デフォルト設定を返す
func DefaultKarnoffelConfig() KarnoffelConfig {
	return KarnoffelConfig{
		CpuDifficulty: KarnoffelCpuDifficultyNormal,
		TargetHands:   KarnoffelDefaultTarget,
	}
}

// Validate 設定値のドメインバリデーション
func (c KarnoffelConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(KarnoffelCpuDifficultyNormal), int(KarnoffelCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("target hands", c.TargetHands, KarnoffelMinTarget, KarnoffelMaxTarget)
}
