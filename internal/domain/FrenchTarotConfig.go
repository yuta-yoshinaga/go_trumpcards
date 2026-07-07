//go:build !js || !wasm || extra

package domain

// FrenchTarotDefenderCnt デクレアラーに対する防御側 (defenders) の人数。
// コイン精算の倍率に用いる (デクレアラー 1 + 防御 3 = 4 人)。
const FrenchTarotDefenderCnt = FrenchTarotPlayerCnt - 1

// FrenchTarotCpuDifficulty CPU の難易度レベル
type FrenchTarotCpuDifficulty int

// French Tarot の CPU 難易度定数
const (
	// FrenchTarotCpuDifficultyEasy 低難易度 (ランダムプレイ)
	FrenchTarotCpuDifficultyEasy FrenchTarotCpuDifficulty = iota
	// FrenchTarotCpuDifficultyNormal 中難易度 (戦略プレイ)
	FrenchTarotCpuDifficultyNormal
	// FrenchTarotCpuDifficultyHard 高難易度 (戦略プレイ)
	FrenchTarotCpuDifficultyHard
)

// FrenchTarotConfig フレンチタロット (French Tarot) のゲーム設定
type FrenchTarotConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty FrenchTarotCpuDifficulty `json:"cd"`
	// TargetDeals マッチを構成するディール数。この回数だけ配り、累積得点最上位が勝者。
	TargetDeals int `json:"td"`
}

// DefaultFrenchTarotConfig デフォルト設定を返す (標準は 4 ディール)。
func DefaultFrenchTarotConfig() FrenchTarotConfig {
	return FrenchTarotConfig{CpuDifficulty: FrenchTarotCpuDifficultyNormal, TargetDeals: FrenchTarotDefaultDeals}
}

// Validate 設定値のドメインバリデーション
func (c FrenchTarotConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(FrenchTarotCpuDifficultyEasy), int(FrenchTarotCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target deals", c.TargetDeals, 1); err != nil {
		return err
	}
	return nil
}
