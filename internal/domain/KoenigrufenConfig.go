//go:build !js || !wasm || extra

package domain

// KoenigrufenCpuDifficulty CPU の難易度レベル
type KoenigrufenCpuDifficulty int

// Königrufen の CPU 難易度定数
const (
	// KoenigrufenCpuDifficultyEasy 低難易度 (ランダムプレイ)
	KoenigrufenCpuDifficultyEasy KoenigrufenCpuDifficulty = iota
	// KoenigrufenCpuDifficultyNormal 中難易度 (戦略プレイ)
	KoenigrufenCpuDifficultyNormal
	// KoenigrufenCpuDifficultyHard 高難易度 (戦略プレイ)
	KoenigrufenCpuDifficultyHard
)

// KoenigrufenConfig ケーニッヒルーフェン (Königrufen) のゲーム設定
type KoenigrufenConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty KoenigrufenCpuDifficulty `json:"cd"`
	// TargetDeals マッチを構成するディール数。この回数だけ配り、累積得点最上位が勝者。
	TargetDeals int `json:"td"`
}

// DefaultKoenigrufenConfig デフォルト設定を返す (標準は 4 ディール)。
func DefaultKoenigrufenConfig() KoenigrufenConfig {
	return KoenigrufenConfig{CpuDifficulty: KoenigrufenCpuDifficultyNormal, TargetDeals: KoenigrufenDefaultDeals}
}

// Validate 設定値のドメインバリデーション
func (c KoenigrufenConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(KoenigrufenCpuDifficultyEasy), int(KoenigrufenCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target deals", c.TargetDeals, 1); err != nil {
		return err
	}
	return nil
}
