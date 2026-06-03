//go:build !js || !wasm || casino

package domain

import "sort"

// 2-7 Triple Draw (Deuce to Seven Lowball) hand evaluation.
//
// The hand is read as an ordinary 5-card poker hand, but the WORST poker hand
// wins. Two rules distinguish it from A-5 lowball (Razz):
//   1. The Ace is ALWAYS high — A-2-3-4-5 is NOT a straight (it is merely
//      "Ace-high"), so the unbeatable nut low is 7-5-4-3-2.
//   2. Straights and flushes DO count: they are full poker hands and therefore
//      among the worst possible holdings for a low.
//
// evalDeuceToSevenHand returns the standard PokerHand* category constant so the
// existing PokerHandNames slice can be reused for display, and
// compareDeuceToSevenCards orders hands so that the lower poker hand wins.

// evalDeuceToSevenHand evaluates a 5-card hand and returns its standard poker
// category (PokerHandHighCard … PokerHandRoyalFlush). Unlike evalFiveCardHand
// it never treats A-2-3-4-5 as a straight (Ace is strictly high).
func evalDeuceToSevenHand(cards []*Card) int {
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

	isFlush := true
	for i := 1; i < 5; i++ {
		if designs[i] != designs[0] {
			isFlush = false
			break
		}
	}

	isStraight := checkDeuceToSevenStraight(values)

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
		// Royal vs plain straight flush only affects the display name; both are
		// the worst possible lows. A-10-J-Q-K is the only Ace straight here.
		if values[0] == 1 {
			return PokerHandRoyalFlush
		}
		return PokerHandStraightFlush
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

// checkDeuceToSevenStraight reports whether the sorted values form a straight
// under 2-7 rules: the Ace is always high, so A-2-3-4-5 is NOT a straight and
// A-10-J-Q-K (broadway) is the only Ace-bearing straight.
func checkDeuceToSevenStraight(sortedValues []int) bool {
	if len(sortedValues) != 5 {
		return false
	}
	if sortedValues[0] == 1 {
		// Only broadway counts when an Ace is present.
		return sortedValues[1] == 10 && sortedValues[2] == 11 &&
			sortedValues[3] == 12 && sortedValues[4] == 13
	}
	for i := 1; i < len(sortedValues); i++ {
		if sortedValues[i] != sortedValues[i-1]+1 {
			return false
		}
	}
	return true
}

// deuceLowCompareKey returns the hand's values arranged for 2-7 tie-breaking:
// grouped by frequency (pairs/trips compare as a block) and ordered high→low
// with Ace = 14. Within a category the hand whose key is lexicographically
// smaller is the stronger (lower) low hand.
func deuceLowCompareKey(cards []*Card) []int {
	freq := make(map[int]int, len(cards))
	for _, c := range cards {
		v := c.GetValue()
		if v == 1 {
			v = 14 // Ace is always high
		}
		freq[v]++
	}
	type group struct{ val, cnt int }
	groups := make([]group, 0, len(freq))
	for v, cnt := range freq {
		groups = append(groups, group{val: v, cnt: cnt})
	}
	// Higher count first (pairs dominate the comparison), then higher value
	// first so the worst card leads.
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].cnt != groups[j].cnt {
			return groups[i].cnt > groups[j].cnt
		}
		return groups[i].val > groups[j].val
	})
	key := make([]int, 0, len(cards))
	for _, g := range groups {
		for k := 0; k < g.cnt; k++ {
			key = append(key, g.val)
		}
	}
	return key
}

// compareDeuceToSevenCards compares two 5-card hands under 2-7 lowball rules.
// Returns -1 if a is the stronger (lower) hand, 1 if b is stronger, 0 on tie.
// Lower poker category wins; on equal category the lexicographically smaller
// frequency-grouped key (Ace = 14) wins.
func compareDeuceToSevenCards(a, b []*Card) int {
	rankA := evalDeuceToSevenHand(a)
	rankB := evalDeuceToSevenHand(b)
	if rankA != rankB {
		if rankA < rankB {
			return -1
		}
		return 1
	}
	keyA := deuceLowCompareKey(a)
	keyB := deuceLowCompareKey(b)
	for i := 0; i < len(keyA) && i < len(keyB); i++ {
		if keyA[i] < keyB[i] {
			return -1
		}
		if keyA[i] > keyB[i] {
			return 1
		}
	}
	return 0
}

// deuceLowStrength returns a 1..4 strength rating used by the CPU to decide
// betting and drawing. Higher is a stronger low. Mirrors the BadugiHand.Size
// scale so the shared CPU style-parameter structure can be reused.
//
//	4 = made pat low, 8-high or better (no pair/straight/flush) — e.g. 7-5-4-3-2
//	3 = made low 9- to jack-high, no pair
//	2 = queen-high or worse, no pair (still salvageable by drawing)
//	1 = any pair or worse (two pair / trips / straight / flush) — drawing dead-ish
func deuceLowStrength(cards []*Card) int {
	if len(cards) != 5 {
		return 1
	}
	rank := evalDeuceToSevenHand(cards)
	if rank != PokerHandHighCard {
		return 1
	}
	high := deuceLowCompareKey(cards)[0] // worst card, Ace = 14
	switch {
	case high <= 8:
		return 4
	case high <= 11:
		return 3
	default:
		return 2
	}
}

// FindPotWinnersDeuceToSeven selects the best (lowest) 2-7 hand among the
// eligible players for pot distribution; equal hands split. Designed to be
// passed to DistributePotsWithWinnerFunc. Each player's GetComparisonCards()
// returns the full 5-card hand.
func FindPotWinnersDeuceToSeven(players []BettingPlayer, eligible []int) []int {
	var bestCards []*Card
	bestInit := false
	var winners []int

	for _, idx := range eligible {
		pl := players[idx]
		if pl.GetFolded() {
			continue
		}
		cards := pl.GetComparisonCards()
		if !bestInit {
			bestCards = cards
			winners = []int{idx}
			bestInit = true
			continue
		}
		switch compareDeuceToSevenCards(cards, bestCards) {
		case -1:
			bestCards = cards
			winners = []int{idx}
		case 0:
			winners = append(winners, idx)
		}
	}
	return winners
}
