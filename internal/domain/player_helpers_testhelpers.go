//go:build test

package domain

// setHandForTest replaces a seat's hand outright. 9 games had this written out
// behind their exported SetHandForTest.
//
// The constraint is spelled `*T` rather than plain handWriter so that a nil
// player can be detected: GetPlayer returns a typed nil for an out-of-range
// index, and a typed nil stored in an interface is not itself nil, so an
// interface parameter would turn the guard into a panic. See issue #5185.
func setHandForTest[T any, P interface {
	*T
	handWriter
}](p P, cards []*Card) {
	if p == nil {
		return
	}
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	for _, c := range cards {
		p.AddCard(c)
	}
}
