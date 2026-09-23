//go:build !js || !wasm || extra

package domain

// TappTarockCpuDifficulty CPU の難易度レベル
type TappTarockCpuDifficulty int

// TappTarock の CPU 難易度定数
const (
	// TappTarockCpuDifficultyEasy 低難易度 (ランダムプレイ)
	TappTarockCpuDifficultyEasy TappTarockCpuDifficulty = iota
	// TappTarockCpuDifficultyNormal 中難易度 (戦略プレイ)
	TappTarockCpuDifficultyNormal
	// TappTarockCpuDifficultyHard 高難易度 (戦略プレイ)
	TappTarockCpuDifficultyHard
)

// TappTarockConfig タップ・タロック (TappTarock) のゲーム設定
type TappTarockConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty TappTarockCpuDifficulty `json:"cd"`
	// TargetDeals マッチを構成するディール数。この回数だけ配り、累積得点最上位が勝者。
	TargetDeals int `json:"td"`
}

// DefaultTappTarockConfig デフォルト設定を返す (標準は 4 ディール)。
func DefaultTappTarockConfig() TappTarockConfig {
	return TappTarockConfig{
		CpuDifficulty: TappTarockCpuDifficultyNormal,
		TargetDeals:   TappTarockDefaultDeals,
	}
}

// Validate 設定値のドメインバリデーション
func (c TappTarockConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(TappTarockCpuDifficultyEasy), int(TappTarockCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateRange("target deals", c.TargetDeals,
		TappTarockMinDeals, TappTarockMaxDeals)
}
