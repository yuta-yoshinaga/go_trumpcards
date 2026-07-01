//go:build !js || !wasm || extra

package domain

// CalabresellaCoalitionSize 連合 (非ソリスト) 側の人数。得点移動の倍率に用いる。
const CalabresellaCoalitionSize = CalabresellaPlayerCnt - 1

// CalabresellaCpuDifficulty CPU の難易度レベル
type CalabresellaCpuDifficulty int

// Calabresella の CPU 難易度定数
const (
	// CalabresellaCpuDifficultyEasy 低難易度 (ランダムプレイ・常にパス)
	CalabresellaCpuDifficultyEasy CalabresellaCpuDifficulty = iota
	// CalabresellaCpuDifficultyNormal 中難易度 (戦略プレイ)
	CalabresellaCpuDifficultyNormal
	// CalabresellaCpuDifficultyHard 高難易度 (戦略プレイ)
	CalabresellaCpuDifficultyHard
)

// CalabresellaConfig カラブレセッラのゲーム設定
type CalabresellaConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty CalabresellaCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積点。いずれかのプレイヤーがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultCalabresellaConfig デフォルト設定を返す (標準は 21 点先取)。
func DefaultCalabresellaConfig() CalabresellaConfig {
	return CalabresellaConfig{CpuDifficulty: CalabresellaCpuDifficultyNormal, TargetPoints: CalabresellaWinTarget}
}

// Validate 設定値のドメインバリデーション
func (c CalabresellaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CalabresellaCpuDifficultyEasy), int(CalabresellaCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
