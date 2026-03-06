package domain

import (
	"fmt"
	"sort"
	"strings"
)

// CardValueDisplayName converts a card value (1-13) to its display label.
func CardValueDisplayName(value int) string {
	switch value {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	default:
		return fmt.Sprintf("%d", value)
	}
}

// ExtractKickers extracts kicker card values from a 5-card hand.
// Returns nil for hands where kickers are not applicable
// (HighCard, Straight, Flush, FullHouse, StraightFlush, RoyalFlush, FiveOfAKind).
// For pair-based hands, returns the non-pair card values sorted descending (Ace=14).
func ExtractKickers(cards []*Card, handRank int) []int {
	switch handRank {
	case PokerHandHighCard, PokerHandStraight, PokerHandFlush,
		PokerHandFullHouse, PokerHandStraightFlush, PokerHandRoyalFlush,
		PokerHandFiveOfAKind:
		return nil
	}

	if len(cards) != 5 {
		return nil
	}

	// Count value frequencies
	freq := make(map[int]int)
	for _, c := range cards {
		freq[c.GetValue()]++
	}

	// Determine the "group" frequency for the hand rank
	var groupFreq int
	switch handRank {
	case PokerHandOnePair:
		groupFreq = 2
	case PokerHandTwoPair:
		groupFreq = 2
	case PokerHandThreeOfAKind:
		groupFreq = 3
	case PokerHandFourOfAKind:
		groupFreq = 4
	}

	// Collect kicker values (cards not part of the hand group)
	kickers := make([]int, 0)
	for val, cnt := range freq {
		if cnt < groupFreq {
			v := val
			if v == 1 {
				v = 14 // Ace high
			}
			// For TwoPair, the remaining single card is the kicker
			// For OnePair, all three singles are kickers
			for i := 0; i < cnt; i++ {
				kickers = append(kickers, v)
			}
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(kickers)))
	return kickers
}

// FormatKickers formats kicker values as a display string (e.g., "A, Q, 10").
func FormatKickers(kickers []int) string {
	if len(kickers) == 0 {
		return ""
	}
	parts := make([]string, len(kickers))
	for i, v := range kickers {
		displayVal := v
		if v == 14 {
			displayVal = 1 // Convert back to card value for display
		}
		parts[i] = CardValueDisplayName(displayVal)
	}
	return strings.Join(parts, ", ")
}
