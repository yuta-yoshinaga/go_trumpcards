//go:build test

package domain

// MarriageCardPoints returns the deadwood points for a card (test helper).
func MarriageCardPoints(card *Card, wildRank int) int {
	return marriageCardPoints(card, wildRank)
}
