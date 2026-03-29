package domain

// OhHellCpuDifficulty CPU の難易度レベル
type OhHellCpuDifficulty int

// OhHellのCPU難易度定数
const (
	// OhHellCpuDifficultyEasy 低難易度
	OhHellCpuDifficultyEasy OhHellCpuDifficulty = iota
	// OhHellCpuDifficultyNormal 中難易度
	OhHellCpuDifficultyNormal
	// OhHellCpuDifficultyHard 高難易度
	OhHellCpuDifficultyHard
)

// OhHellScoringVariant スコアリング方式
type OhHellScoringVariant int

// OhHellのスコアリング定数
const (
	// OhHellScoringStandard 正確なビッド = 10+bid, 外れ = 0
	OhHellScoringStandard OhHellScoringVariant = iota
	// OhHellScoringPenalty 正確なビッド = 10+bid, 外れ = -|差分|
	OhHellScoringPenalty
)

// OhHellRoundDirection ラウンド進行方向
type OhHellRoundDirection int

// OhHellのラウンド方向定数
const (
	// OhHellRoundDownOnly 手札枚数が減少のみ (max→1)
	OhHellRoundDownOnly OhHellRoundDirection = iota
	// OhHellRoundDownAndUp 手札枚数が減少後増加 (max→1→max)
	OhHellRoundDownAndUp
)

// OhHellConfig オー・ヘルゲーム設定
type OhHellConfig struct {
	CpuDifficulty  OhHellCpuDifficulty  `json:"cd"`
	MaxHandSize    int                  `json:"mh"` // 最大手札枚数 (デフォルト10)
	ScoringVariant OhHellScoringVariant `json:"sv"`
	RoundDirection OhHellRoundDirection `json:"rd"`
}

// DefaultOhHellConfig デフォルト設定を返す
func DefaultOhHellConfig() OhHellConfig {
	return OhHellConfig{
		CpuDifficulty:  OhHellCpuDifficultyNormal,
		MaxHandSize:    10,
		ScoringVariant: OhHellScoringStandard,
		RoundDirection: OhHellRoundDownAndUp,
	}
}

// Validate 設定値のドメインバリデーション
func (c OhHellConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(OhHellCpuDifficultyEasy), int(OhHellCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("max hand size", c.MaxHandSize, 1, 13); err != nil {
		return err
	}
	if err := ValidateRange("scoring variant", int(c.ScoringVariant), int(OhHellScoringStandard), int(OhHellScoringPenalty)); err != nil {
		return err
	}
	if err := ValidateRange("round direction", int(c.RoundDirection), int(OhHellRoundDownOnly), int(OhHellRoundDownAndUp)); err != nil {
		return err
	}
	return nil
}
