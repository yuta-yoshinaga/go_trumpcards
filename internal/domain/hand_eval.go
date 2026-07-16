package domain

import "sort"

// 3-card poker hand rank constants (descending order).
// Note: In 3-card poker, Three of a Kind ranks ABOVE Straight.
const (
	ThreeCardHandHighCard      = 1
	ThreeCardHandPair          = 2
	ThreeCardHandFlush         = 3
	ThreeCardHandStraight      = 4
	ThreeCardHandThreeOfAKind  = 5
	ThreeCardHandStraightFlush = 6
)

// ThreeCardHandNames maps 3-card hand ranks to display names.
var ThreeCardHandNames = []string{
	"", // 0 unused
	"High Card",
	"Pair",
	"Flush",
	"Straight",
	"Three of a Kind",
	"Straight Flush",
}

// evalThreeCardHand evaluates a 3-card poker hand and returns its rank.
func evalThreeCardHand(cards []*Card) int {
	if len(cards) != 3 {
		return ThreeCardHandHighCard
	}

	values := make([]int, 3)
	designs := make([]int, 3)
	for i, c := range cards {
		values[i] = c.GetValue()
		designs[i] = c.GetDesign()
	}
	sort.Ints(values)

	isFlush := designs[0] == designs[1] && designs[1] == designs[2]
	isStraight := checkThreeCardStraight(values)

	// Value frequency count
	valueCounts := make(map[int]int)
	for _, v := range values {
		valueCounts[v]++
	}

	if isFlush && isStraight {
		return ThreeCardHandStraightFlush
	}
	// Three of a Kind ranks above Straight in 3-card poker
	for _, cnt := range valueCounts {
		if cnt == 3 {
			return ThreeCardHandThreeOfAKind
		}
	}
	if isStraight {
		return ThreeCardHandStraight
	}
	if isFlush {
		return ThreeCardHandFlush
	}
	for _, cnt := range valueCounts {
		if cnt == 2 {
			return ThreeCardHandPair
		}
	}
	return ThreeCardHandHighCard
}

// checkThreeCardStraight checks if 3 sorted values form a straight.
// Valid wraps: A-2-3 and Q-K-A.
func checkThreeCardStraight(sortedValues []int) bool {
	if len(sortedValues) != 3 {
		return false
	}
	// A-2-3
	if sortedValues[0] == 1 && sortedValues[1] == 2 && sortedValues[2] == 3 {
		return true
	}
	// Q-K-A
	if sortedValues[0] == 1 && sortedValues[1] == 12 && sortedValues[2] == 13 {
		return true
	}
	// Normal consecutive
	return sortedValues[1] == sortedValues[0]+1 && sortedValues[2] == sortedValues[1]+1
}

// threeCardHandHighValues returns card values sorted for comparison,
// treating Ace as 14 (high).
func threeCardHandHighValues(cards []*Card) []int {
	vals := make([]int, len(cards))
	for i, c := range cards {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		vals[i] = v
	}
	sort.Sort(sort.Reverse(sort.IntSlice(vals)))
	return vals
}

// threeCardStraightHighCard returns the high card value for a 3-card straight,
// using special handling for A-2-3 (high card = 3) and Q-K-A (high card = 14).
func threeCardStraightHighCard(cards []*Card) int {
	values := make([]int, 3)
	for i, c := range cards {
		values[i] = c.GetValue()
	}
	sort.Ints(values)
	// A-2-3: high card is 3
	if values[0] == 1 && values[1] == 2 && values[2] == 3 {
		return 3
	}
	// Q-K-A: high card is 14 (Ace high)
	if values[0] == 1 && values[1] == 12 && values[2] == 13 {
		return 14
	}
	return values[2]
}

// compareThreeCardHands compares two 3-card poker hands.
// Returns 1 if a wins, -1 if b wins, 0 if tie.
func compareThreeCardHands(a, b []*Card) int {
	rankA := evalThreeCardHand(a)
	rankB := evalThreeCardHand(b)
	if rankA > rankB {
		return 1
	}
	if rankA < rankB {
		return -1
	}

	// Same rank — compare by kickers
	switch rankA {
	case ThreeCardHandStraight, ThreeCardHandStraightFlush:
		highA := threeCardStraightHighCard(a)
		highB := threeCardStraightHighCard(b)
		if highA > highB {
			return 1
		}
		if highA < highB {
			return -1
		}
		return 0
	case ThreeCardHandThreeOfAKind:
		// All three same value; compare that value
		vA := a[0].GetValue()
		if vA == 1 {
			vA = 14
		}
		vB := b[0].GetValue()
		if vB == 1 {
			vB = 14
		}
		if vA > vB {
			return 1
		}
		if vA < vB {
			return -1
		}
		return 0
	default:
		// High card, Flush, Pair: compare each card descending
		valsA := threeCardHandHighValues(a)
		valsB := threeCardHandHighValues(b)
		// For pairs, sort so pair value comes first
		if rankA == ThreeCardHandPair {
			valsA = threeCardPairSortedValues(a)
			valsB = threeCardPairSortedValues(b)
		}
		for i := 0; i < len(valsA); i++ {
			if valsA[i] > valsB[i] {
				return 1
			}
			if valsA[i] < valsB[i] {
				return -1
			}
		}
		return 0
	}
}

// threeCardPairSortedValues returns values sorted with pair value first, then kicker.
func threeCardPairSortedValues(cards []*Card) []int {
	freq := make(map[int]int)
	for _, c := range cards {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		freq[v]++
	}
	var pairVal, kickerVal int
	for v, cnt := range freq {
		if cnt == 2 {
			pairVal = v
		} else {
			kickerVal = v
		}
	}
	return []int{pairVal, kickerVal}
}

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

// evalRazzHand evaluates a 5-card hand for A-5 lowball (Razz).
// Straights and flushes do NOT count against the player.
// Only pair-type categories matter. Returns the same PokerHand* constants.
func evalRazzHand(cards []*Card) int {
	if len(cards) != 5 {
		return PokerHandHighCard
	}

	valueCounts := make(map[int]int)
	for _, c := range cards {
		valueCounts[c.GetValue()]++
	}
	counts := make([]int, 0, len(valueCounts))
	for _, cnt := range valueCounts {
		counts = append(counts, cnt)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(counts)))

	if counts[0] == 4 {
		return PokerHandFourOfAKind
	}
	if len(counts) >= 2 && counts[0] == 3 && counts[1] == 2 {
		return PokerHandFullHouse
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

// compareRazzCards compares two Razz hands (A-5 lowball: Ace=1, lower wins).
// Returns -1 if a is stronger (lower), 1 if b is stronger, 0 if tie.
func compareRazzCards(a, b []*Card) int {
	aVals := razzCardValues(a)
	bVals := razzCardValues(b)
	sort.Sort(sort.Reverse(sort.IntSlice(aVals)))
	sort.Sort(sort.Reverse(sort.IntSlice(bVals)))
	for i := 0; i < len(aVals) && i < len(bVals); i++ {
		if aVals[i] < bVals[i] {
			return -1
		}
		if aVals[i] > bVals[i] {
			return 1
		}
	}
	return 0
}

// razzCardValues returns card values for Razz comparison (Ace=1, Joker=0).
func razzCardValues(cards []*Card) []int {
	vals := make([]int, len(cards))
	for i, c := range cards {
		if c.GetDesign() == CardDesignJoker {
			vals[i] = 0
		} else {
			vals[i] = c.GetValue() // Ace stays as 1
		}
	}
	return vals
}

// SevenCardStudRazzBestLow returns the strongest 5-card Razz low from cards and
// whether it is complete. With fewer than 5 cards the low cannot be made yet, so
// the cards are returned sorted ascending (Ace low) for a progress display and
// complete is false. This is a pure, non-mutating read used by the CUI.
func SevenCardStudRazzBestLow(cards []*Card) (best []*Card, complete bool) {
	if len(cards) < 5 {
		sorted := make([]*Card, len(cards))
		copy(sorted, cards)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].GetValue() < sorted[j].GetValue()
		})
		return sorted, false
	}
	bestRank := -1
	for _, combo := range combinations(cards, 5) {
		rank := evalRazzHand(combo)
		if bestRank == -1 || rank < bestRank || (rank == bestRank && compareRazzCards(combo, best) < 0) {
			bestRank = rank
			best = append([]*Card(nil), combo...)
		}
	}
	return best, true
}
