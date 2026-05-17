package domain

import "sort"

// 4-card poker hand rank constants (ascending order).
// In Four Card Poker the order is: High Card < Pair < Two Pair < Straight <
// Flush < Three of a Kind < Straight Flush < Four of a Kind. Notably, Flush
// outranks Straight because four cards make straights more common than flushes.
const (
	FourCardHandHighCard      = 1
	FourCardHandPair          = 2
	FourCardHandTwoPair       = 3
	FourCardHandStraight      = 4
	FourCardHandFlush         = 5
	FourCardHandThreeOfAKind  = 6
	FourCardHandStraightFlush = 7
	FourCardHandFourOfAKind   = 8
)

// FourCardHandNames maps 4-card hand ranks to display names.
var FourCardHandNames = []string{
	"", // 0 unused
	"High Card",
	"Pair",
	"Two Pair",
	"Straight",
	"Flush",
	"Three of a Kind",
	"Straight Flush",
	"Four of a Kind",
}

// evalFourCardHand evaluates a 4-card poker hand and returns its rank.
func evalFourCardHand(cards []*Card) int {
	if len(cards) != 4 {
		return FourCardHandHighCard
	}

	values := make([]int, 4)
	designs := make([]int, 4)
	for i, c := range cards {
		values[i] = c.GetValue()
		designs[i] = c.GetDesign()
	}
	sort.Ints(values)

	isFlush := designs[0] == designs[1] && designs[1] == designs[2] && designs[2] == designs[3]
	isStraight := checkFourCardStraight(values)

	valueCounts := make(map[int]int)
	for _, v := range values {
		valueCounts[v]++
	}
	counts := make([]int, 0, len(valueCounts))
	for _, cnt := range valueCounts {
		counts = append(counts, cnt)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(counts)))

	if counts[0] == 4 {
		return FourCardHandFourOfAKind
	}
	if isFlush && isStraight {
		return FourCardHandStraightFlush
	}
	if counts[0] == 3 {
		return FourCardHandThreeOfAKind
	}
	if isFlush {
		return FourCardHandFlush
	}
	if isStraight {
		return FourCardHandStraight
	}
	if len(counts) >= 2 && counts[0] == 2 && counts[1] == 2 {
		return FourCardHandTwoPair
	}
	if counts[0] == 2 {
		return FourCardHandPair
	}
	return FourCardHandHighCard
}

// checkFourCardStraight checks if 4 sorted values form a straight.
// Valid wraps: A-2-3-4 (wheel) and J-Q-K-A (broadway).
func checkFourCardStraight(sortedValues []int) bool {
	if len(sortedValues) != 4 {
		return false
	}
	if sortedValues[0] == 1 && sortedValues[1] == 2 && sortedValues[2] == 3 && sortedValues[3] == 4 {
		return true
	}
	if sortedValues[0] == 1 && sortedValues[1] == 11 && sortedValues[2] == 12 && sortedValues[3] == 13 {
		return true
	}
	for i := 1; i < len(sortedValues); i++ {
		if sortedValues[i] != sortedValues[i-1]+1 {
			return false
		}
	}
	return true
}

// fourCardHandHighValues returns card values sorted descending with Ace=14.
func fourCardHandHighValues(cards []*Card) []int {
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

// fourCardStraightHighCard returns the high card value for a 4-card straight,
// with A-2-3-4 = 4 and J-Q-K-A = 14.
func fourCardStraightHighCard(cards []*Card) int {
	values := make([]int, 4)
	for i, c := range cards {
		values[i] = c.GetValue()
	}
	sort.Ints(values)
	if values[0] == 1 && values[1] == 2 && values[2] == 3 && values[3] == 4 {
		return 4
	}
	if values[0] == 1 && values[1] == 11 && values[2] == 12 && values[3] == 13 {
		return 14
	}
	return values[3]
}

// fourCardPairSortedValues returns descending values with pair value first, then kickers.
func fourCardPairSortedValues(cards []*Card) []int {
	freq := make(map[int]int)
	for _, c := range cards {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		freq[v]++
	}
	pairVal := 0
	kickers := make([]int, 0, 2)
	for v, cnt := range freq {
		if cnt == 2 {
			pairVal = v
		} else {
			kickers = append(kickers, v)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(kickers)))
	out := append([]int{pairVal}, kickers...)
	return out
}

// fourCardTwoPairSortedValues returns the two pair values descending.
func fourCardTwoPairSortedValues(cards []*Card) []int {
	freq := make(map[int]int)
	for _, c := range cards {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		freq[v]++
	}
	pairs := make([]int, 0, 2)
	for v, cnt := range freq {
		if cnt == 2 {
			pairs = append(pairs, v)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(pairs)))
	return pairs
}

// fourCardTripsKicker returns the trip value and kicker (descending).
func fourCardTripsKicker(cards []*Card) (trip, kicker int) {
	freq := make(map[int]int)
	for _, c := range cards {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		freq[v]++
	}
	for v, cnt := range freq {
		switch cnt {
		case 3:
			trip = v
		case 1:
			kicker = v
		}
	}
	return trip, kicker
}

// fourCardQuadValue returns the quad value (Ace=14).
func fourCardQuadValue(cards []*Card) int {
	v := cards[0].GetValue()
	if v == 1 {
		v = 14
	}
	return v
}

// compareFourCardHands compares two 4-card poker hands.
// Returns 1 if a wins, -1 if b wins, 0 if tie.
func compareFourCardHands(a, b []*Card) int {
	rankA := evalFourCardHand(a)
	rankB := evalFourCardHand(b)
	if rankA > rankB {
		return 1
	}
	if rankA < rankB {
		return -1
	}

	switch rankA {
	case FourCardHandStraight, FourCardHandStraightFlush:
		highA := fourCardStraightHighCard(a)
		highB := fourCardStraightHighCard(b)
		return cmpInt(highA, highB)
	case FourCardHandFourOfAKind:
		return cmpInt(fourCardQuadValue(a), fourCardQuadValue(b))
	case FourCardHandThreeOfAKind:
		tripA, kickA := fourCardTripsKicker(a)
		tripB, kickB := fourCardTripsKicker(b)
		if c := cmpInt(tripA, tripB); c != 0 {
			return c
		}
		return cmpInt(kickA, kickB)
	case FourCardHandTwoPair:
		pa := fourCardTwoPairSortedValues(a)
		pb := fourCardTwoPairSortedValues(b)
		for i := 0; i < len(pa) && i < len(pb); i++ {
			if c := cmpInt(pa[i], pb[i]); c != 0 {
				return c
			}
		}
		return 0
	case FourCardHandPair:
		va := fourCardPairSortedValues(a)
		vb := fourCardPairSortedValues(b)
		for i := 0; i < len(va) && i < len(vb); i++ {
			if c := cmpInt(va[i], vb[i]); c != 0 {
				return c
			}
		}
		return 0
	default:
		// High Card, Flush
		va := fourCardHandHighValues(a)
		vb := fourCardHandHighValues(b)
		for i := 0; i < len(va) && i < len(vb); i++ {
			if c := cmpInt(va[i], vb[i]); c != 0 {
				return c
			}
		}
		return 0
	}
}

func cmpInt(a, b int) int {
	if a > b {
		return 1
	}
	if a < b {
		return -1
	}
	return 0
}

// pickBestFour selects the strongest 4-card hand out of n cards (n >= 4).
// Generates C(n,4) combinations and picks the highest via compareFourCardHands.
func pickBestFour(cards []*Card) []*Card {
	n := len(cards)
	if n <= 4 {
		out := make([]*Card, len(cards))
		copy(out, cards)
		return out
	}
	var best []*Card
	indices := make([]int, 4)
	var pick func(start, depth int)
	pick = func(start, depth int) {
		if depth == 4 {
			combo := []*Card{cards[indices[0]], cards[indices[1]], cards[indices[2]], cards[indices[3]]}
			if best == nil || compareFourCardHands(combo, best) > 0 {
				best = combo
			}
			return
		}
		for i := start; i < n; i++ {
			indices[depth] = i
			pick(i+1, depth+1)
		}
	}
	pick(0, 0)
	return best
}

// pickBestFourFromFive is a convenience wrapper for the 5-card player hand.
func pickBestFourFromFive(cards []*Card) []*Card {
	return pickBestFour(cards)
}
