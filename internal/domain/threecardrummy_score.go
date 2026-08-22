//go:build !js || !wasm || casino

package domain

import "sort"

// Three Card Rummy の採点。**低いほど強い。**
//
// ブラックジャック系と違い、目標は「点を作らない」こと。3 枚の合計が低いほど
// 良く、0 点が最強。既存の `ThreeCard` (Three Card Poker) が役の強さで比べるのに
// 対し、ここは合計点の低さで比べるので、評価軸が根本的に違う。

const (
	// ThreeCardRummyFaceValue は絵札 (J/Q/K) の点数。
	ThreeCardRummyFaceValue = 10
	// ThreeCardRummyAceValue はエースの点数。**常に 1** —— 11 にはならない。
	ThreeCardRummyAceValue = 1
	// ThreeCardRummyPerfectScore は「役」が成立した手の点数 (最強)。
	ThreeCardRummyPerfectScore = 0
	// ThreeCardRummyDealerQualifyMax はディーラーがクオリファイする上限。
	// これ以下なら勝負が成立する。
	ThreeCardRummyDealerQualifyMax = 20
)

// threeCardRummyCardValue は 1 枚の点数を返す。
func threeCardRummyCardValue(c *Card) int {
	switch v := c.GetValue(); {
	case v == 1:
		return ThreeCardRummyAceValue
	case v >= 11: // J / Q / K
		return ThreeCardRummyFaceValue
	default:
		return v
	}
}

// ThreeCardRummyScore は 3 枚の点数を返す。**低いほど強い。**
//
// **同スートの連番 3 枚、または同ランク 3 枚は 0 点。** 実質的な最強手で、
// 合計を素直に足すだけの実装ではこの 2 つが「最悪の手」になってしまう
// (K-K-K は 30 点、Q-K-A は 21 点)。
func ThreeCardRummyScore(cards []*Card) int {
	if threeCardRummyIsMeld(cards) {
		return ThreeCardRummyPerfectScore
	}
	total := 0
	for _, c := range cards {
		if c == nil {
			continue
		}
		total += threeCardRummyCardValue(c)
	}
	return total
}

// threeCardRummyIsMeld は同ランク 3 枚か、同スートの連番 3 枚かを返す。
//
// 3 枚ちょうど揃っていない手 (配り途中や nil 混じり) は役ではない。**ここで
// 弾かないと 0 点 = 最強**として扱われてしまうので、長さと nil を先に見る。
func threeCardRummyIsMeld(cards []*Card) bool {
	if len(cards) != ThreeCardRummyHandSize {
		return false
	}
	for _, c := range cards {
		if c == nil {
			return false
		}
	}
	return threeCardRummyIsSet(cards) || threeCardRummyIsRun(cards)
}

// threeCardRummyIsSet は同ランク 3 枚か。
func threeCardRummyIsSet(cards []*Card) bool {
	v := cards[0].GetValue()
	for _, c := range cards[1:] {
		if c.GetValue() != v {
			return false
		}
	}
	return true
}

// threeCardRummyIsRun は同スートの連番 3 枚か。
//
// **A は下端にも上端にも付く。** A-2-3 も Q-K-A も連番として認める ——
// ラミーの並びは 1 本の輪ではないが、どちらの端も伝統的に有効な組。
func threeCardRummyIsRun(cards []*Card) bool {
	suit := cards[0].GetDesign()
	for _, c := range cards[1:] {
		if c.GetDesign() != suit {
			return false
		}
	}
	vals := make([]int, 0, ThreeCardRummyHandSize)
	for _, c := range cards {
		vals = append(vals, c.GetValue())
	}
	sort.Ints(vals)
	// A を 1 として扱った並び。
	if vals[1] == vals[0]+1 && vals[2] == vals[1]+1 {
		return true
	}
	// A を 14 として扱った並び (Q-K-A)。
	if vals[0] == 1 && vals[1] == 12 && vals[2] == 13 {
		return true
	}
	return false
}
