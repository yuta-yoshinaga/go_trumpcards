package domain

import "sort"

// evalFiveCardHand evaluates a 5-card poker hand and returns its rank.
func evalFiveCardHand(cards []*Card) int {
	if len(cards) != 5 {
		return PokerHandHighCard
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
	isStraight := checkStraightValues(values)

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
			return PokerHandRoyalFlush
		}
		return PokerHandStraightFlush
	}
	if counts[0] == 5 {
		return PokerHandFiveOfAKind
	}
	if counts[0] == 4 {
		return PokerHandFourOfAKind
	}
	if len(counts) >= 2 && counts[0] == 3 && counts[1] == 2 {
		return PokerHandFullHouse
	}
	if isFlush {
		return PokerHandFlush
	}
	if isStraight {
		return PokerHandStraight
	}
	if counts[0] == 3 {
		return PokerHandThreeOfAKind
	}
	if len(counts) >= 2 && counts[0] == 2 && counts[1] == 2 {
		return PokerHandTwoPair
	}
	if counts[0] == 2 {
		return PokerHandOnePair
	}
	return PokerHandHighCard
}

// checkStraightValues checks if sorted values form a straight.
func checkStraightValues(sortedValues []int) bool {
	// A-2-3-4-5 (ホイール)
	if sortedValues[0] == 1 && sortedValues[1] == 2 &&
		sortedValues[2] == 3 && sortedValues[3] == 4 && sortedValues[4] == 5 {
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

// checkRoyalStraightValues checks if sorted values form a royal straight.
func checkRoyalStraightValues(sortedValues []int) bool {
	return len(sortedValues) == 5 &&
		sortedValues[0] == 1 &&
		sortedValues[1] == 10 &&
		sortedValues[2] == 11 &&
		sortedValues[3] == 12 &&
		sortedValues[4] == 13
}
