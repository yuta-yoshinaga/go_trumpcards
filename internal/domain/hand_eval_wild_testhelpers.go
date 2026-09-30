//go:build test

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
