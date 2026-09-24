//go:build !js || !wasm || extra4

package domain

// PutMinMatchTarget マッチ目標点の下限
const PutMinMatchTarget = 1

// PutMaxMatchTarget マッチ目標点の上限
const PutMaxMatchTarget = 60

// PutDefaultMatchTarget マッチ目標点の既定値 (put a 15)
const PutDefaultMatchTarget = 15

// PutConfig プットゲーム設定
type PutConfig struct {
	// MatchTarget この点数に最初に到達したプレイヤーがマッチに勝利する
	MatchTarget int `json:"mt"`
}

// DefaultPutConfig デフォルト設定を返す
func DefaultPutConfig() PutConfig {
	return PutConfig{
		MatchTarget: PutDefaultMatchTarget,
	}
}

// Validate 設定値のドメインバリデーション
func (c PutConfig) Validate() error {
	return ValidateRange("Match target", c.MatchTarget, PutMinMatchTarget, PutMaxMatchTarget)
}

// normalized 不正値を既定値に丸めた設定を返す (Reset 時の安全弁)。
func (c PutConfig) normalized() PutConfig {
	if c.MatchTarget < PutMinMatchTarget || c.MatchTarget > PutMaxMatchTarget {
		c.MatchTarget = PutDefaultMatchTarget
	}
	return c
}
