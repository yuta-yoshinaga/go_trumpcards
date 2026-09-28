//go:build test

package domain

// SetDeckForTest replaces the deck order for deterministic tests.
func (b *Basset) SetDeckForTest(cards []*Card) {
	b.trumpCards = NewTrumpCardsWithSuits(0, []int{})
	b.trumpCards.deck, b.trumpCards.deckCnt = cards, len(cards)
	b.trumpCards.deckInit()
}
