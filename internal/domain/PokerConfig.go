package domain

import "fmt"

// PokerCpuCountMin PokerCpuCountMax CPU プレイヤー数の有効範囲
const (
	PokerCpuCountMin = 1 // CPU プレイヤー最小数
	PokerCpuCountMax = 3 // CPU プレイヤー最大数
)

// PokerJokerCountMax ジョーカー枚数最大
const PokerJokerCountMax = 2

// PokerPlayStyle CPUプレイスタイル
type PokerPlayStyle int

// Pokerのプレイスタイル定数
const (
	PokerStyleConservative PokerPlayStyle = iota // 保守的
	PokerStyleBalanced                           // バランス
	PokerStyleAggressive                         // 攻撃的
	PokerStyleBluffer                            // ブラフ重視
)

// PokerPlayStyleNames プレイスタイル名
var PokerPlayStyleNames = []string{
	"Conservative",
	"Balanced",
	"Aggressive",
	"Bluffer",
}

// PokerConfig ポーカー設定
type PokerConfig struct {
	InitChips    int              // 初期チップ
	Ante         int              // アンティ
	MinBet       int              // 最小ベット
	CpuCount     int              // CPU数 (1-3)
	JokerCount   int              // ジョーカー枚数 (0-2)
	BettingLimit BettingLimitType // ベッティングリミット
	IsLowball    bool             // 2-7 Lowball モード
	CpuMetaAI    bool             // メタAI: セッション内学習
}

// DefaultPokerConfig デフォルト設定
func DefaultPokerConfig() PokerConfig {
	return PokerConfig{
		InitChips:  1000,
		Ante:       10,
		MinBet:     10,
		CpuCount:   3,
		JokerCount: 0,
	}
}

// Validate 設定値のドメインバリデーション
func (c PokerConfig) Validate() error {
	if c.BettingLimit < BettingLimitFixed || c.BettingLimit > BettingLimitNoLimit {
		return fmt.Errorf("betting limit must be %d-%d, got %d", int(BettingLimitFixed), int(BettingLimitNoLimit), int(c.BettingLimit))
	}
	if c.CpuCount < PokerCpuCountMin || c.CpuCount > PokerCpuCountMax {
		return fmt.Errorf("CPU player count must be %d-%d, got %d", PokerCpuCountMin, PokerCpuCountMax, c.CpuCount)
	}
	if c.JokerCount < 0 || c.JokerCount > PokerJokerCountMax {
		return fmt.Errorf("joker count must be 0-%d, got %d", PokerJokerCountMax, c.JokerCount)
	}
	return nil
}

// pokerCpuStyleParams CPU意思決定パラメータ
type pokerCpuStyleParams struct {
	aggressive bool // true=Aggressive, false=Passive
	bluffRate  int  // ブラフ率(%)

	// 第1ベッティングラウンド
	firstBetThreshold  int // ハンドランク >= this → ベット
	firstBetMult       int // ベット = MinBet * this
	firstFoldThreshold int // ハンドランク <= this → フォールド候補
	firstCallMaxMult   int // コール額 > MinBet * this → フォールド (Passive only)

	// 第2ベッティングラウンド
	secondBetThreshold  int // ハンドランク >= this → ベット/レイズ
	secondBetMult       int // ベット = MinBet * this
	secondFoldThreshold int // ハンドランク <= this → フォールド候補
	secondCallMaxMult   int // コール額 > MinBet * this → フォールド

	// 交換枚数読み補正
	exchangeReadWeight int // 相手の交換枚数少ない → 強い手警戒 (0-100)

	// スタンドパットブラフ
	standPatBluffRate int // 弱い手でも交換しない確率(%)
}

// pokerStyleParamsMap スタイルごとのパラメータ
var pokerStyleParamsMap = map[PokerPlayStyle]pokerCpuStyleParams{
	PokerStyleConservative: {
		aggressive:          false,
		bluffRate:           5,
		firstBetThreshold:   PokerHandTwoPair,
		firstBetMult:        1,
		firstFoldThreshold:  PokerHandHighCard,
		firstCallMaxMult:    2,
		secondBetThreshold:  PokerHandTwoPair,
		secondBetMult:       2,
		secondFoldThreshold: PokerHandHighCard,
		secondCallMaxMult:   2,
		exchangeReadWeight:  80,
		standPatBluffRate:   0,
	},
	PokerStyleBalanced: {
		aggressive:          false,
		bluffRate:           15,
		firstBetThreshold:   PokerHandOnePair,
		firstBetMult:        1,
		firstFoldThreshold:  PokerHandHighCard,
		firstCallMaxMult:    3,
		secondBetThreshold:  PokerHandOnePair,
		secondBetMult:       2,
		secondFoldThreshold: PokerHandHighCard,
		secondCallMaxMult:   3,
		exchangeReadWeight:  50,
		standPatBluffRate:   5,
	},
	PokerStyleAggressive: {
		aggressive:          true,
		bluffRate:           25,
		firstBetThreshold:   PokerHandOnePair,
		firstBetMult:        2,
		firstFoldThreshold:  PokerHandHighCard,
		firstCallMaxMult:    5,
		secondBetThreshold:  PokerHandOnePair,
		secondBetMult:       3,
		secondFoldThreshold: PokerHandHighCard,
		secondCallMaxMult:   5,
		exchangeReadWeight:  30,
		standPatBluffRate:   10,
	},
	PokerStyleBluffer: {
		aggressive:          true,
		bluffRate:           40,
		firstBetThreshold:   PokerHandHighCard,
		firstBetMult:        2,
		firstFoldThreshold:  PokerHandHighCard,
		firstCallMaxMult:    4,
		secondBetThreshold:  PokerHandHighCard,
		secondBetMult:       3,
		secondFoldThreshold: PokerHandHighCard,
		secondCallMaxMult:   4,
		exchangeReadWeight:  20,
		standPatBluffRate:   20,
	},
}
