//go:build !js || !wasm || extra6

package domain

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
	// TargetScore は勝利に要る通算点。
	TargetScore int `json:"ts"`
	// AllowSchmeiss は「投げ」を許すか。
	AllowSchmeiss bool `json:"as"`
}

// DefaultKlaberjassConfig デフォルト設定を返す
func DefaultKlaberjassConfig() KlaberjassConfig {
	return KlaberjassConfig{
		TargetScore:   KlaberjassTargetScoreDefault,
		AllowSchmeiss: true,
	}
}

// Validate 設定値のドメインバリデーション
func (c KlaberjassConfig) Validate() error {
	return ValidateRange("target score", c.TargetScore, KlaberjassTargetScoreMin, KlaberjassTargetScoreMax)
}
