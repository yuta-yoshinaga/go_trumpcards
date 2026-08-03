//go:build !js || !wasm || casino

package domain

import "sort"

// サイドベット種類定数
const (
	BJSideBetPerfectPairs = 1
	BJSideBet21Plus3      = 2
)

// Perfect Pairs 結果定数
const (
	BJPPNone        = 0 // ペアなし
	BJPPMixedPair   = 1 // ミックスペア（異色同数）
	BJPPColoredPair = 2 // カラードペア（同色同数）
	BJPPPerfectPair = 3 // パーフェクトペア（同柄同数）
)

// Perfect Pairs 配当倍率
const (
	BJPPMixedPairPayout   = 6
	BJPPColoredPairPayout = 12
	BJPPPerfectPairPayout = 25
)

// 21+3 結果定数
const (
	BJT3None          = 0 // なし
	BJT3Flush         = 1 // フラッシュ
	BJT3Straight      = 2 // ストレート
	BJT3ThreeOfAKind  = 3 // スリーオブアカインド
	BJT3StraightFlush = 4 // ストレートフラッシュ
	BJT3SuitedTrips   = 5 // スーテッドトリップス
)

// 21+3 配当倍率
const (
	BJT3FlushPayout         = 5
	BJT3StraightPayout      = 10
	BJT3ThreeOfAKindPayout  = 30
	BJT3StraightFlushPayout = 40
	BJT3SuitedTripsPayout   = 100
)

// BJSideBetResult サイドベット結果
type BJSideBetResult struct {
	BetType    int    `json:"bt"`
	ResultType int    `json:"rt"`
	ResultName string `json:"rn"`
	BetAmount  int    `json:"ba"`
	Payout     int    `json:"po"`
}

// BetTypeName ベット種別名を返す
func (r *BJSideBetResult) BetTypeName() string {
	switch r.BetType {
	case BJSideBetPerfectPairs:
		return "Perfect Pairs"
	case BJSideBet21Plus3:
		// Displayed name, not an identifier — "21+3" is a live Japanese
		// trademark (登録6752649 / 6785367, Galaxy Gaming, classes 28 & 41,
		// which cover providing online games), so the shipped label describes
		// the bet instead of naming it. See TRADEMARKS.md.
		return "Poker Hand Bonus"
	default:
		return "Unknown"
	}
}

// EvaluatePerfectPairs プレイヤーの最初の2枚でPerfect Pairsを判定
func EvaluatePerfectPairs(card1, card2 *Card) (resultType int, resultName string) {
	if !isSameValue(card1, card2) {
		return BJPPNone, ""
	}
	if isSameSuit(card1, card2) {
		return BJPPPerfectPair, "Perfect Pair"
	}
	if isSameColor(card1, card2) {
		return BJPPColoredPair, "Colored Pair"
	}
	return BJPPMixedPair, "Mixed Pair"
}

// Evaluate21Plus3 プレイヤーの2枚+ディーラーのアップカードで21+3を判定
func Evaluate21Plus3(card1, card2, dealerUpcard *Card) (resultType int, resultName string) {
	if isThreeOfAKind3(card1, card2, dealerUpcard) && isFlush3(card1, card2, dealerUpcard) {
		return BJT3SuitedTrips, "Suited Trips"
	}
	if isStraight3(card1, card2, dealerUpcard) && isFlush3(card1, card2, dealerUpcard) {
		return BJT3StraightFlush, "Straight Flush"
	}
	if isThreeOfAKind3(card1, card2, dealerUpcard) {
		return BJT3ThreeOfAKind, "Three of a Kind"
	}
	if isStraight3(card1, card2, dealerUpcard) {
		return BJT3Straight, "Straight"
	}
	if isFlush3(card1, card2, dealerUpcard) {
		return BJT3Flush, "Flush"
	}
	return BJT3None, ""
}

// PerfectPairsPayout Perfect Pairsの配当倍率を返す
func PerfectPairsPayout(resultType int) int {
	switch resultType {
	case BJPPPerfectPair:
		return BJPPPerfectPairPayout
	case BJPPColoredPair:
		return BJPPColoredPairPayout
	case BJPPMixedPair:
		return BJPPMixedPairPayout
	default:
		return 0
	}
}

// TwentyOnePlus3Payout 21+3の配当倍率を返す
func TwentyOnePlus3Payout(resultType int) int {
	switch resultType {
	case BJT3SuitedTrips:
		return BJT3SuitedTripsPayout
	case BJT3StraightFlush:
		return BJT3StraightFlushPayout
	case BJT3ThreeOfAKind:
		return BJT3ThreeOfAKindPayout
	case BJT3Straight:
		return BJT3StraightPayout
	case BJT3Flush:
		return BJT3FlushPayout
	default:
		return 0
	}
}

// isSameSuit 同じスートか
func isSameSuit(a, b *Card) bool {
	return a.GetDesign() == b.GetDesign()
}

// isSameColor 同じ色か（SPADE+CLOVER=黒, HEART+DIAMOND=赤）
func isSameColor(a, b *Card) bool {
	return cardColor(a) == cardColor(b)
}

// cardColor カードの色を返す（0=黒, 1=赤）
func cardColor(c *Card) int {
	d := c.GetDesign()
	if d == CardDesignHeart || d == CardDesignDiamond {
		return 1
	}
	return 0
}

// isSameValue 同じ数字か
func isSameValue(a, b *Card) bool {
	return a.GetValue() == b.GetValue()
}

// isFlush3 3枚が同じスートか
func isFlush3(a, b, c *Card) bool {
	return a.GetDesign() == b.GetDesign() && b.GetDesign() == c.GetDesign()
}

// isStraight3 3枚がストレート（連番）か（A-2-3 と Q-K-A のラップも許可）
func isStraight3(a, b, c *Card) bool {
	vals := []int{a.GetValue(), b.GetValue(), c.GetValue()}
	sort.Ints(vals)
	// 通常ストレート: v[0]+1==v[1] && v[1]+1==v[2]
	if vals[0]+1 == vals[1] && vals[1]+1 == vals[2] {
		return true
	}
	// A-ラップ: Q(12)-K(13)-A(1) => sorted: 1, 12, 13
	if vals[0] == 1 && vals[1] == 12 && vals[2] == 13 {
		return true
	}
	return false
}

// isThreeOfAKind3 3枚が同じ数字か
func isThreeOfAKind3(a, b, c *Card) bool {
	return a.GetValue() == b.GetValue() && b.GetValue() == c.GetValue()
}
