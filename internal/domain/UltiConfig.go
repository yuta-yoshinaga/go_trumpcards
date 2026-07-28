//go:build !js || !wasm || extra3

package domain

// UltiCoalitionSize 連合 (非デクレアラー) 側の人数。コイン移動の倍率に用いる。
const UltiCoalitionSize = UltiPlayerCnt - 1

// UltiCpuDifficulty CPU の難易度レベル
type UltiCpuDifficulty int

// Ulti の CPU 難易度定数
const (
	// UltiCpuDifficultyEasy 低難易度 (ランダムプレイ)
	UltiCpuDifficultyEasy UltiCpuDifficulty = iota
	// UltiCpuDifficultyNormal 中難易度 (戦略プレイ)
	UltiCpuDifficultyNormal
	// UltiCpuDifficultyHard 高難易度 (戦略プレイ)
	UltiCpuDifficultyHard
)

// UltiConfig ウルティ (Ulti) のゲーム設定
type UltiConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty UltiCpuDifficulty `json:"cd"`
	// TargetRounds マッチを構成するディール数。この回数だけ配り、累積コイン最上位が勝者。
	TargetRounds int `json:"tr"`
}

// DefaultUltiConfig デフォルト設定を返す (標準は 5 ディール)。
func DefaultUltiConfig() UltiConfig {
	return UltiConfig{CpuDifficulty: UltiCpuDifficultyNormal, TargetRounds: UltiWinRounds}
}

// Validate 設定値のドメインバリデーション
func (c UltiConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(UltiCpuDifficultyEasy), int(UltiCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
