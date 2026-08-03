//go:build !js || !wasm || extra3

package domain

// KlaberjassCpuDifficulty CPU の難易度レベル
type KlaberjassCpuDifficulty int

// Klaberjass の CPU 難易度定数
const (
	// KlaberjassCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	KlaberjassCpuDifficultyNormal KlaberjassCpuDifficulty = iota
)

// KlaberjassTargetScoreDefault は既定の目標点。
//
// **501 点。**issue #4395 は目標点を書いていない。
const KlaberjassTargetScoreDefault = 501

// KlaberjassTargetScoreMin / KlaberjassTargetScoreMax は目標点の範囲。
const (
	KlaberjassTargetScoreMin = 100
	KlaberjassTargetScoreMax = 1000
)

// KlaberjassConfig クラバーヤスのゲーム設定
type KlaberjassConfig struct {
	CpuDifficulty KlaberjassCpuDifficulty `json:"cd"`
	// TargetScore は勝利に要る通算点。
	TargetScore int `json:"ts"`
	// AllowSchmeiss は「投げ」を許すか。
	AllowSchmeiss bool `json:"as"`
}

// DefaultKlaberjassConfig デフォルト設定を返す
func DefaultKlaberjassConfig() KlaberjassConfig {
	return KlaberjassConfig{
		CpuDifficulty: KlaberjassCpuDifficultyNormal,
		TargetScore:   KlaberjassTargetScoreDefault,
		AllowSchmeiss: true,
	}
}

// Validate 設定値のドメインバリデーション
func (c KlaberjassConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(KlaberjassCpuDifficultyNormal), int(KlaberjassCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, KlaberjassTargetScoreMin, KlaberjassTargetScoreMax)
}
