package domain

// card_color.go holds card-colour helpers shared across categories. They live
// in an untagged (core) file — not in a per-category game file — so the
// per-worker build-tag split (#2126) keeps them available in every Cloudflare
// Worker. For example Nertz (classic worker) reuses isAlternateColor, which
// originally lived in freecell_solver.go (solo).

// isBlack reports whether a card belongs to a black suit (spade or club).
func isBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

// isAlternateColor reports whether two cards are of opposite colours.
func isAlternateColor(card1, card2 *Card) bool {
	return isBlack(card1) != isBlack(card2)
}
