//go:build !js || !wasm || extra3

package domain

// OmbreCoalitionSize 連合 (非オンブル) 側の人数。得点移動の倍率に用いる。
const OmbreCoalitionSize = OmbrePlayerCnt - 1

// OmbreCpuDifficulty CPU の難易度レベル
type OmbreCpuDifficulty int

// Ombre の CPU 難易度定数
const (
	// OmbreCpuDifficultyEasy 低難易度 (ランダムプレイ・常にパス)
	OmbreCpuDifficultyEasy OmbreCpuDifficulty = iota
	// OmbreCpuDifficultyNormal 中難易度 (戦略プレイ)
	OmbreCpuDifficultyNormal
	// OmbreCpuDifficultyHard 高難易度 (戦略プレイ)
	OmbreCpuDifficultyHard
)

// OmbreConfig オンブル (Ombre) のゲーム設定
type OmbreConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty OmbreCpuDifficulty `json:"cd"`
	// TargetRounds マッチを構成するディール数。この回数だけ配り、累積点最上位が勝者。
	TargetRounds int `json:"tr"`
}

// DefaultOmbreConfig デフォルト設定を返す (標準は 5 ディール)。
func DefaultOmbreConfig() OmbreConfig {
	return OmbreConfig{CpuDifficulty: OmbreCpuDifficultyNormal, TargetRounds: OmbreWinRounds}
}

// Validate 設定値のドメインバリデーション
func (c OmbreConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(OmbreCpuDifficultyEasy), int(OmbreCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
