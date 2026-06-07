//go:build !js || !wasm || casino

package domain

import "sort"

// ショートデックハンドランク定数 (Flush > FullHouse)
const (
	ShortDeckHandHighCard      = 0
	ShortDeckHandOnePair       = 1
	ShortDeckHandTwoPair       = 2
	ShortDeckHandThreeOfAKind  = 3
	ShortDeckHandStraight      = 4
	ShortDeckHandFullHouse     = 5
	ShortDeckHandFlush         = 6
	ShortDeckHandFourOfAKind   = 7
	ShortDeckHandStraightFlush = 8
	ShortDeckHandRoyalFlush    = 9
)

// ShortDeckHandNames ショートデックハンド名
var ShortDeckHandNames = []string{
	"High Card",
	"One Pair",
	"Two Pair",
	"Three of a Kind",
	"Straight",
	"Full House",
	"Flush",
	"Four of a Kind",
	"Straight Flush",
	"Royal Flush",
}

// ShortDeckValues ショートデックで使用するカード値 (A,6,7,8,9,10,J,Q,K)

// evalShortDeckFiveCardHand ショートデック用5枚ハンド評価
// Flush > FullHouse, A-6-7-8-9が最弱ストレート
func evalShortDeckFiveCardHand(cards []*Card) int {
	if len(cards) != 5 {
		return ShortDeckHandHighCard
	}

	values := make([]int, 5)
	designs := make([]int, 5)
	for i, c := range cards {
		values[i] = c.GetValue()
		designs[i] = c.GetDesign()
	}
	sort.Ints(values)

	// フラッシュチェック
	isFlush := true
	for i := 1; i < 5; i++ {
		if designs[i] != designs[0] {
			isFlush = false
			break
		}
	}

	// ストレートチェック
	isStraight := checkShortDeckStraightValues(values)

	// カード値の出現回数カウント
	valueCounts := make(map[int]int)
	for _, v := range values {
		valueCounts[v]++
	}
	counts := make([]int, 0, len(valueCounts))
	for _, c := range valueCounts {
		counts = append(counts, c)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(counts)))

	if isFlush && isStraight {
		if checkRoyalStraightValues(values) {
			return ShortDeckHandRoyalFlush
		}
		return ShortDeckHandStraightFlush
	}
	if counts[0] == 4 {
		return ShortDeckHandFourOfAKind
	}
	// ショートデック: Flush > FullHouse
	if isFlush {
		return ShortDeckHandFlush
	}
	if len(counts) >= 2 && counts[0] == 3 && counts[1] == 2 {
		return ShortDeckHandFullHouse
	}
	if isStraight {
		return ShortDeckHandStraight
	}
	if counts[0] == 3 {
		return ShortDeckHandThreeOfAKind
	}
	if len(counts) >= 2 && counts[0] == 2 && counts[1] == 2 {
		return ShortDeckHandTwoPair
	}
	if counts[0] == 2 {
		return ShortDeckHandOnePair
	}
	return ShortDeckHandHighCard
}

// checkShortDeckStraightValues ショートデック用ストレート判定
// A-6-7-8-9 (ショートデックホイール) と A-10-J-Q-K (ブロードウェイ) を特別扱い
func checkShortDeckStraightValues(sortedValues []int) bool {
	// A-6-7-8-9 (ショートデックホイール: 最弱ストレート)
	if sortedValues[0] == 1 && sortedValues[1] == 6 &&
		sortedValues[2] == 7 && sortedValues[3] == 8 && sortedValues[4] == 9 {
		return true
	}
	// A-10-J-Q-K (ブロードウェイ)
	if sortedValues[0] == 1 && sortedValues[1] == 10 &&
		sortedValues[2] == 11 && sortedValues[3] == 12 && sortedValues[4] == 13 {
		return true
	}
	// 通常のストレート
	for i := 1; i < len(sortedValues); i++ {
		if sortedValues[i] != sortedValues[i-1]+1 {
			return false
		}
	}
	return true
}

// isShortDeckWheelHand ショートデックホイール (A-6-7-8-9) かどうか判定
func isShortDeckWheelHand(cards []*Card) bool {
	if len(cards) != 5 {
		return false
	}
	vals := make([]int, 5)
	for i, c := range cards {
		vals[i] = c.GetValue()
	}
	sort.Ints(vals)
	return vals[0] == 1 && vals[1] == 6 && vals[2] == 7 && vals[3] == 8 && vals[4] == 9
}

// compareShortDeckHighCardsSlice ショートデック用ハイカード比較
// ホイールがA-6-7-8-9になる以外はcompareHighCardsSliceと同じ
func compareShortDeckHighCardsSlice(a, b []*Card) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	aWheel := isShortDeckWheelHand(a)
	bWheel := isShortDeckWheelHand(b)
	aVals := make([]int, len(a))
	bVals := make([]int, len(b))
	for i, c := range a {
		v := c.GetValue()
		if v == 1 && !aWheel {
			v = 14
		}
		aVals[i] = v
	}
	for i, c := range b {
		v := c.GetValue()
		if v == 1 && !bWheel {
			v = 14
		}
		bVals[i] = v
	}
	aTB := tieBreakValues(aVals)
	bTB := tieBreakValues(bVals)
	for i := 0; i < len(aTB) && i < len(bTB); i++ {
		if aTB[i] > bTB[i] {
			return 1
		}
		if aTB[i] < bTB[i] {
			return -1
		}
	}
	return 0
}
