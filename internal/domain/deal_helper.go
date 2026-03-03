package domain

// cardAdder is implemented by any player type that can receive cards.
type cardAdder interface {
	AddCard(card *Card)
}

// dealAllCards draws all cards from the deck and distributes them
// round-robin to players.
func dealAllCards[T cardAdder](deck *TrumpCards, players []T) {
	idx := 0
	for {
		card := deck.DrawCard()
		if card == nil {
			break
		}
		players[idx%len(players)].AddCard(card)
		idx++
	}
}
