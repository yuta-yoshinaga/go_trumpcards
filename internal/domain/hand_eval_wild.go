package domain

// countWilds returns the number of wild cards in the hand.
func countWilds(cards []*Card, isWild func(*Card) bool) int {
	if isWild == nil {
		return 0
	}
	count := 0
	for _, c := range cards {
		if isWild(c) {
			count++
		}
	}
	return count
}

// evalWildHand evaluates a 5-card poker hand that may contain wild cards.
// It returns the best achievable hand rank and whether any wild cards were used.
// If isWild is nil, all cards are treated as non-wild.
func evalWildHand(cards []*Card, isWild func(*Card) bool) (bestRank int, usedWilds bool) {
	if len(cards) != 5 {
		return PokerHandHighCard, false
	}

	// No wild card function: standard evaluation
	if isWild == nil {
		return evalFiveCardHand(cards), false
	}

	// Separate wild and non-wild cards
	var wilds []*Card
	var nonWilds []*Card
	for _, c := range cards {
		if isWild(c) {
			wilds = append(wilds, c)
		} else {
			nonWilds = append(nonWilds, c)
		}
	}

	numWilds := len(wilds)
	if numWilds == 0 {
		return evalFiveCardHand(cards), false
	}

	// 4+ wilds: best possible is Five of a Kind
	if numWilds >= 4 {
		return PokerHandFiveOfAKind, true
	}

	// **Five of a Kind is decidable without searching at all.** If some value
	// already appears often enough that the wilds can top it up to five, every
	// wild takes that value and the hand is the best there is. This is the
	// common "pair plus wilds" shape, and skipping the search for it is what
	// keeps the worst case off the request budget rather than merely smaller.
	if maxValueCount(nonWilds)+numWilds >= 5 {
		return PokerHandFiveOfAKind, true
	}

	// Enumerate substitutions for the wild cards over the 52-card alphabet.
	//
	// **Wild cards are interchangeable, so only the multiset matters.**
	// Enumerating ordered tuples evaluates every assignment numWilds! times:
	// with three wilds that is 52^3 = 140,608 five-card evaluations instead of
	// the 24,804 distinct ones, and the hand takes over half a second — long
	// past what a Cloudflare Worker request is allowed to spend. Walking
	// non-decreasing index tuples covers exactly the same set of hands.
	bestRank = PokerHandHighCard
	allSubs := make([]*Card, numWilds)
	alphabet := wildSubstitutionAlphabet(nonWilds)

	var enumerate func(depth, from int)
	enumerate = func(depth, from int) {
		// **The ceiling with wilds is Five of a Kind, not a Royal Flush.**
		// Stopping at the royal flush alone leaves the common "pair plus
		// wilds" hand grinding through the whole alphabet after it has
		// already found the best hand there is.
		if bestRank == PokerHandFiveOfAKind {
			return
		}
		if depth == numWilds {
			// Build the 5-card hand: nonWilds + substitutions
			hand := make([]*Card, 0, 5)
			hand = append(hand, nonWilds...)
			hand = append(hand, allSubs...)
			rank := evalFiveCardHand(hand)
			if rank > bestRank {
				bestRank = rank
			}
			return
		}
		for i := from; i < len(alphabet); i++ {
			if bestRank == PokerHandFiveOfAKind {
				return
			}
			allSubs[depth] = alphabet[i]
			enumerate(depth+1, i)
		}
	}
	enumerate(0, 0)

	return bestRank, true
}

// wildSubstitutionAlphabet returns the cards a wild may usefully stand in for,
// given the non-wild cards it will sit beside.
//
// **Most of the 52-card alphabet is redundant, and which part depends on the
// suits already on the table.** A five-card hand is a flush, straight flush or
// royal flush only when all five share a suit, so:
//
//   - If the non-wilds are already of two or more suits, no wild can make the
//     hand a flush of any kind. Every remaining rank — pairs, trips, straights,
//     full houses, quads — reads only the values, so the wild's suit cannot
//     change the answer and one suit's worth of values covers every outcome.
//     That is 13 candidates instead of 52, which is 64x fewer assignments for
//     three wilds.
//   - If the non-wilds do share a suit, a flush is still reachable, so that
//     suit has to stay. One other suit is enough to represent "not the flush
//     suit", since beyond the flush the suit is again irrelevant.
//
// An empty non-wild set (five wilds) never reaches here; four or more wilds
// return Five of a Kind outright.
func wildSubstitutionAlphabet(nonWilds []*Card) []*Card {
	suits := []int{CardDesignSpade}
	if flushSuit := commonSuit(nonWilds); flushSuit > 0 {
		other := CardDesignSpade
		if other == flushSuit {
			other = CardDesignHeart
		}
		suits = []int{flushSuit, other}
	}
	out := make([]*Card, 0, len(suits)*13)
	for _, suit := range suits {
		for val := 1; val <= 13; val++ {
			out = append(out, NewCard(suit, val, false))
		}
	}
	return out
}

// commonSuit returns the suit every card shares, or -1 when they differ.
// An empty slice has no common suit to speak of and reports -1.
func commonSuit(cards []*Card) int {
	suit := -1
	for _, c := range cards {
		if c == nil {
			continue
		}
		if suit == -1 {
			suit = c.GetDesign()
			continue
		}
		if c.GetDesign() != suit {
			return -1
		}
	}
	return suit
}

// maxValueCount returns how often the most frequent card value appears.
func maxValueCount(cards []*Card) int {
	if len(cards) == 0 {
		return 0
	}
	counts := make(map[int]int, len(cards))
	best := 0
	for _, c := range cards {
		if c == nil {
			continue
		}
		counts[c.GetValue()]++
		best = max(best, counts[c.GetValue()])
	}
	return best
}
