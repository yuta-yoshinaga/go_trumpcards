package domain

import (
	"slices"
	"sort"
)

// Big Two card value strength: 3 < 4 < ... < K < A < 2
// Suit strength: ♦(4) < ♣(2) < ♥(3) < ♠(1)
// Combined strength = valueStrength * 4 + suitStrength

// bigTwoValueStrength returns the value strength (3=0, 4=1, ..., K=10, A=11, 2=12).
func bigTwoValueStrength(v int) int {
	if v == 2 {
		return 12
	}
	if v == 1 {
		return 11
	}
	return v - 3
}

// bigTwoSuitStrength returns the suit strength (♦=0, ♣=1, ♥=2, ♠=3).
func bigTwoSuitStrength(design int) int {
	switch design {
	case CardDesignDiamond:
		return 0
	case CardDesignClover:
		return 1
	case CardDesignHeart:
		return 2
	case CardDesignSpade:
		return 3
	default:
		return 0
	}
}

// BigTwoCardStrength returns the combined strength of a card.
func BigTwoCardStrength(card *Card) int {
	return bigTwoValueStrength(card.GetValue())*4 + bigTwoSuitStrength(card.GetDesign())
}

// BigTwoPlayType プレイの種類
type BigTwoPlayType int

// BigTwoPlayType定数
const (
	BigTwoPlayInvalid       BigTwoPlayType = 0
	BigTwoPlaySingle        BigTwoPlayType = 1
	BigTwoPlayPair          BigTwoPlayType = 2
	BigTwoPlayTriple        BigTwoPlayType = 3
	BigTwoPlayStraight      BigTwoPlayType = 4
	BigTwoPlayFlush         BigTwoPlayType = 5
	BigTwoPlayFullHouse     BigTwoPlayType = 6
	BigTwoPlayFourOfAKind   BigTwoPlayType = 7
	BigTwoPlayStraightFlush BigTwoPlayType = 8
)

// bigTwoClassifyPlay classifies a set of cards into its Big Two play type.
func bigTwoClassifyPlay(cards []*Card) BigTwoPlayType {
	n := len(cards)
	switch n {
	case 1:
		return BigTwoPlaySingle
	case 2:
		if cards[0].GetValue() == cards[1].GetValue() {
			return BigTwoPlayPair
		}
		return BigTwoPlayInvalid
	case 3:
		if cards[0].GetValue() == cards[1].GetValue() && cards[1].GetValue() == cards[2].GetValue() {
			return BigTwoPlayTriple
		}
		return BigTwoPlayInvalid
	case 5:
		return bigTwoClassify5Cards(cards)
	default:
		return BigTwoPlayInvalid
	}
}

// bigTwoClassify5Cards classifies a 5-card combination.
func bigTwoClassify5Cards(cards []*Card) BigTwoPlayType {
	values := make([]int, 5)
	designs := make([]int, 5)
	for i, c := range cards {
		values[i] = c.GetValue()
		designs[i] = c.GetDesign()
	}
	sort.Ints(values)

	isFlush := designs[0] == designs[1] && designs[1] == designs[2] &&
		designs[2] == designs[3] && designs[3] == designs[4]
	isStraight := bigTwoCheckStraight(values)

	freq := make(map[int]int)
	for _, v := range values {
		freq[v]++
	}
	counts := make([]int, 0, len(freq))
	for _, c := range freq {
		counts = append(counts, c)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(counts)))

	if isFlush && isStraight {
		return BigTwoPlayStraightFlush
	}
	if counts[0] == 4 {
		return BigTwoPlayFourOfAKind
	}
	if counts[0] == 3 && len(counts) >= 2 && counts[1] == 2 {
		return BigTwoPlayFullHouse
	}
	if isFlush {
		return BigTwoPlayFlush
	}
	if isStraight {
		return BigTwoPlayStraight
	}
	return BigTwoPlayInvalid
}

// bigTwoCheckStraight checks if 5 sorted values form a straight.
// In Big Two, 2 cannot be part of a straight (it is the highest single card).
// Valid straights: 3-4-5-6-7 through 10-J-Q-K-A, plus A-2-3-4-5 is NOT valid.
func bigTwoCheckStraight(sortedValues []int) bool {
	// 2 cannot be in a straight in standard Big Two
	if slices.Contains(sortedValues, 2) {
		return false
	}
	// 10-J-Q-K-A (A acts as high)
	if sortedValues[0] == 1 && sortedValues[1] == 10 &&
		sortedValues[2] == 11 && sortedValues[3] == 12 && sortedValues[4] == 13 {
		return true
	}
	// Normal consecutive (A=1 cannot be low end in Big Two's most common variant)
	// But allow A as high: we already checked 10-J-Q-K-A above
	// If A appears elsewhere, it's not a straight
	if sortedValues[0] == 1 {
		return false
	}
	for i := 1; i < 5; i++ {
		if sortedValues[i] != sortedValues[i-1]+1 {
			return false
		}
	}
	return true
}

// bigTwoPlayStrength returns a comparable strength value for a play.
// Higher is stronger. Returns -1 for invalid plays.
func bigTwoPlayStrength(cards []*Card, playType BigTwoPlayType) int {
	switch playType {
	case BigTwoPlaySingle:
		return BigTwoCardStrength(cards[0])
	case BigTwoPlayPair, BigTwoPlayTriple:
		return bigTwoGroupStrength(cards)
	case BigTwoPlayStraight:
		return bigTwoStraightStrength(cards)
	case BigTwoPlayFlush:
		return bigTwoFlushStrength(cards)
	case BigTwoPlayFullHouse:
		return bigTwoFullHouseStrength(cards)
	case BigTwoPlayFourOfAKind:
		return bigTwoFourOfAKindStrength(cards)
	case BigTwoPlayStraightFlush:
		return bigTwoStraightFlushStrength(cards)
	default:
		return -1
	}
}

// bigTwoGroupStrength returns strength for pairs/triples (by highest card).
func bigTwoGroupStrength(cards []*Card) int {
	maxStr := 0
	for _, c := range cards {
		s := BigTwoCardStrength(c)
		if s > maxStr {
			maxStr = s
		}
	}
	return maxStr
}

// bigTwoStraightStrength returns strength for a straight.
// Compared by highest card value, then by highest card's suit.
func bigTwoStraightStrength(cards []*Card) int {
	highCard := bigTwoHighCardOfStraight(cards)
	return BigTwoCardStrength(highCard)
}

// bigTwoHighCardOfStraight returns the highest card of a straight.
func bigTwoHighCardOfStraight(cards []*Card) *Card {
	sorted := make([]*Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return bigTwoValueStrength(sorted[i].GetValue()) < bigTwoValueStrength(sorted[j].GetValue())
	})
	return sorted[len(sorted)-1]
}

// bigTwoFlushStrength returns strength for a flush.
// Compared by suit first, then by highest card within that suit.
func bigTwoFlushStrength(cards []*Card) int {
	suitStr := bigTwoSuitStrength(cards[0].GetDesign())
	highVal := 0
	for _, c := range cards {
		v := bigTwoValueStrength(c.GetValue())
		if v > highVal {
			highVal = v
		}
	}
	return suitStr*13 + highVal
}

// bigTwoFullHouseStrength returns strength for a full house (by triple's value).
func bigTwoFullHouseStrength(cards []*Card) int {
	freq := make(map[int]int)
	for _, c := range cards {
		freq[c.GetValue()]++
	}
	tripleVal := 0
	for v, cnt := range freq {
		if cnt == 3 {
			tripleVal = v
			break
		}
	}
	return bigTwoValueStrength(tripleVal)
}

// bigTwoFourOfAKindStrength returns strength for four of a kind (by quad's value).
func bigTwoFourOfAKindStrength(cards []*Card) int {
	freq := make(map[int]int)
	for _, c := range cards {
		freq[c.GetValue()]++
	}
	quadVal := 0
	for v, cnt := range freq {
		if cnt == 4 {
			quadVal = v
			break
		}
	}
	return bigTwoValueStrength(quadVal)
}

// bigTwoStraightFlushStrength returns strength for a straight flush.
func bigTwoStraightFlushStrength(cards []*Card) int {
	return bigTwoStraightStrength(cards)
}

// bigTwoIsPlayable checks if cards can be played on the table.
func bigTwoIsPlayable(cards []*Card, tableCards []*Card, tablePlayType BigTwoPlayType) bool {
	playType := bigTwoClassifyPlay(cards)
	if playType == BigTwoPlayInvalid {
		return false
	}

	if tableCards == nil {
		return true
	}

	// 5-card combo hierarchy: Straight < Flush < FullHouse < FourOfAKind < StraightFlush
	if len(cards) == 5 && len(tableCards) == 5 {
		if playType > tablePlayType {
			return true
		}
		if playType < tablePlayType {
			return false
		}
		// Same type: compare strength
		return bigTwoPlayStrength(cards, playType) > bigTwoPlayStrength(tableCards, tablePlayType)
	}

	// For singles, pairs, triples: must be same count and same type
	if len(cards) != len(tableCards) {
		return false
	}
	if playType != tablePlayType {
		return false
	}
	return bigTwoPlayStrength(cards, playType) > bigTwoPlayStrength(tableCards, tablePlayType)
}
