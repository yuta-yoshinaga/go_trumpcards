//go:build !js || !wasm || solo

package domain

// ChinchonCpuDifficulty CPU の難易度レベル
type ChinchonCpuDifficulty int

// ChinchonのCPU難易度定数
const (
	// ChinchonCpuDifficultyEasy 低難易度
	ChinchonCpuDifficultyEasy ChinchonCpuDifficulty = iota
	// ChinchonCpuDifficultyNormal 中難易度
	ChinchonCpuDifficultyNormal
	// ChinchonCpuDifficultyHard 高難易度
	ChinchonCpuDifficultyHard
)

// ChinchonConfig チンチョンゲーム設定
type ChinchonConfig struct {
	CpuDifficulty    ChinchonCpuDifficulty `json:"cd"`
	PlayerCount      int                   `json:"pc"` // プレイヤー数 (2-4)
	KnockThreshold   int                   `json:"kt"` // ノック可能なデッドウッド上限
	EliminationLimit int                   `json:"el"` // 累積点がこの値を超えたプレイヤーは脱落
}

// DefaultChinchonConfig デフォルト設定を返す。
//
// 4人 (人間1 + CPU3)、ノック閾値5点、脱落上限100点。チンチョン (同スート7連続) は
// 設定に依らず常に即時ゲーム勝利となる。
func DefaultChinchonConfig() ChinchonConfig {
	return ChinchonConfig{
		CpuDifficulty:    ChinchonCpuDifficultyNormal,
		PlayerCount:      4,
		KnockThreshold:   5,
		EliminationLimit: 100,
	}
}

// Validate 設定値のドメインバリデーション
func (c ChinchonConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ChinchonCpuDifficultyEasy), int(ChinchonCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("player count", c.PlayerCount, 2, 4); err != nil {
		return err
	}
	// ノック閾値は 0 (ジン相当) 〜 手札6枚分のデッドウッド上限まで許容する。
	if err := ValidateRange("knock threshold", c.KnockThreshold, 0, 60); err != nil {
		return err
	}
	if err := ValidateRange("elimination limit", c.EliminationLimit, 1, 1000); err != nil {
		return err
	}
	return nil
}
