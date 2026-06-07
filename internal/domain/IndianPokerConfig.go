//go:build !js || !wasm || casino

package domain

// IndianPokerConfig インディアンポーカー設定
type IndianPokerConfig struct {
	Ante         int              // アンティ
	InitChips    int              // 初期チップ
	BettingLimit BettingLimitType // ベッティングリミット
	CpuMetaAI    bool             // メタAI: セッション内学習
}

// DefaultIndianPokerConfig デフォルト設定
func DefaultIndianPokerConfig() IndianPokerConfig {
	return IndianPokerConfig{
		Ante:         10,
		InitChips:    1000,
		BettingLimit: BettingLimitNoLimit,
		CpuMetaAI:    true,
	}
}

// Validate 設定値のドメインバリデーション
func (c IndianPokerConfig) Validate() error {
	if err := ValidateMin("ante", c.Ante, 1); err != nil {
		return err
	}
	if err := ValidateMin("init chips", c.InitChips, 1); err != nil {
		return err
	}
	if err := ValidateRange("betting limit", int(c.BettingLimit), int(BettingLimitFixed), int(BettingLimitNoLimit)); err != nil {
		return err
	}
	return nil
}

// indianPokerDefaultCpuStyles インディアンポーカーCPUスタイル (1 human + 3 CPU)
var indianPokerDefaultCpuStyles = []HoldemPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP}

// NewIndianPokerPlayers インディアンポーカーのプレイヤースライスを生成する (1 human + 3 CPU)
func NewIndianPokerPlayers() []*IndianPokerPlayer {
	players := make([]*IndianPokerPlayer, 0, 4)
	players = append(players, NewIndianPokerPlayer(true, HoldemStyleTAG))
	for _, s := range indianPokerDefaultCpuStyles {
		players = append(players, NewIndianPokerPlayer(false, s))
	}
	return players
}
