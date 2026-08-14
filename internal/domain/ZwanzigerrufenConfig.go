//go:build !js || !wasm || extra

package domain

// ZwanzigerrufenCpuDifficulty CPU の難易度レベル
type ZwanzigerrufenCpuDifficulty int

// Zwanzigerrufen の CPU 難易度定数
const (
	// ZwanzigerrufenCpuDifficultyEasy 低難易度 (ランダムプレイ)
	ZwanzigerrufenCpuDifficultyEasy ZwanzigerrufenCpuDifficulty = iota
	// ZwanzigerrufenCpuDifficultyNormal 中難易度 (戦略プレイ)
	ZwanzigerrufenCpuDifficultyNormal
	// ZwanzigerrufenCpuDifficultyHard 高難易度 (戦略プレイ)
	ZwanzigerrufenCpuDifficultyHard
)

// ZwanzigerrufenConfig ツヴァンツィガールーフェン (Zwanzigerrufen) のゲーム設定
type ZwanzigerrufenConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty ZwanzigerrufenCpuDifficulty `json:"cd"`
	// TargetDeals マッチを構成するディール数。この回数だけ配り、累積得点最上位が勝者。
	TargetDeals int `json:"td"`
}

// DefaultZwanzigerrufenConfig デフォルト設定を返す (標準は 4 ディール)。
func DefaultZwanzigerrufenConfig() ZwanzigerrufenConfig {
	return ZwanzigerrufenConfig{
		CpuDifficulty: ZwanzigerrufenCpuDifficultyNormal,
		TargetDeals:   ZwanzigerrufenDefaultDeals,
	}
}

// Validate 設定値のドメインバリデーション
func (c ZwanzigerrufenConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(ZwanzigerrufenCpuDifficultyEasy), int(ZwanzigerrufenCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateRange("target deals", c.TargetDeals,
		ZwanzigerrufenMinDeals, ZwanzigerrufenMaxDeals)
}
