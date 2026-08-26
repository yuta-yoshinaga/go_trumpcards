//go:build !js || !wasm || extra4

package domain

// PutCpuDifficulty CPU の難易度レベル
type PutCpuDifficulty int

// PutのCPU難易度定数
const (
	// PutCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	PutCpuDifficultyNormal PutCpuDifficulty = iota
)

// PutMinMatchTarget マッチ目標点の下限
const PutMinMatchTarget = 1

// PutMaxMatchTarget マッチ目標点の上限
const PutMaxMatchTarget = 60

// PutDefaultMatchTarget マッチ目標点の既定値 (put a 15)
const PutDefaultMatchTarget = 15

// PutConfig プットゲーム設定
type PutConfig struct {
	CpuDifficulty PutCpuDifficulty `json:"cd"`
	// MatchTarget この点数に最初に到達したプレイヤーがマッチに勝利する
	MatchTarget int `json:"mt"`
}

// DefaultPutConfig デフォルト設定を返す
func DefaultPutConfig() PutConfig {
	return PutConfig{
		CpuDifficulty: PutCpuDifficultyNormal,
		MatchTarget:   PutDefaultMatchTarget,
	}
}

// Validate 設定値のドメインバリデーション
func (c PutConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(PutCpuDifficultyNormal), int(PutCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("Match target", c.MatchTarget, PutMinMatchTarget, PutMaxMatchTarget)
}

// normalized 不正値を既定値に丸めた設定を返す (Reset 時の安全弁)。
func (c PutConfig) normalized() PutConfig {
	if c.MatchTarget < PutMinMatchTarget || c.MatchTarget > PutMaxMatchTarget {
		c.MatchTarget = PutDefaultMatchTarget
	}
	return c
}
