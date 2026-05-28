package domain

// TrucoCpuDifficulty CPU の難易度レベル
type TrucoCpuDifficulty int

// TrucoのCPU難易度定数
const (
	// TrucoCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	TrucoCpuDifficultyNormal TrucoCpuDifficulty = iota
)

// TrucoMinMatchTarget マッチ目標点の下限
const TrucoMinMatchTarget = 1

// TrucoMaxMatchTarget マッチ目標点の上限
const TrucoMaxMatchTarget = 60

// TrucoDefaultMatchTarget マッチ目標点の既定値 (truco a 15)
const TrucoDefaultMatchTarget = 15

// TrucoConfig トゥルコゲーム設定
type TrucoConfig struct {
	CpuDifficulty TrucoCpuDifficulty `json:"cd"`
	// MatchTarget この点数に最初に到達したプレイヤーがマッチに勝利する
	MatchTarget int `json:"mt"`
}

// DefaultTrucoConfig デフォルト設定を返す
func DefaultTrucoConfig() TrucoConfig {
	return TrucoConfig{
		CpuDifficulty: TrucoCpuDifficultyNormal,
		MatchTarget:   TrucoDefaultMatchTarget,
	}
}

// Validate 設定値のドメインバリデーション
func (c TrucoConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(TrucoCpuDifficultyNormal), int(TrucoCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("Match target", c.MatchTarget, TrucoMinMatchTarget, TrucoMaxMatchTarget)
}

// normalized 不正値を既定値に丸めた設定を返す (Reset 時の安全弁)。
func (c TrucoConfig) normalized() TrucoConfig {
	if c.MatchTarget < TrucoMinMatchTarget || c.MatchTarget > TrucoMaxMatchTarget {
		c.MatchTarget = TrucoDefaultMatchTarget
	}
	return c
}
