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

	// Enumerate all possible substitutions for wild cards
	// Each wild can be any of 4 suits x 13 values = 52 possibilities
	bestRank = PokerHandHighCard
	allSubs := make([]*Card, numWilds)

	var enumerate func(depth int)
	enumerate = func(depth int) {
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
		for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
			for val := 1; val <= 13; val++ {
				if bestRank == PokerHandRoyalFlush {
					return // early exit: can't improve beyond Royal Flush
				}
				allSubs[depth] = NewCard(suit, val, false)
				enumerate(depth + 1)
			}
		}
	}
	enumerate(0)

	return bestRank, true
}
