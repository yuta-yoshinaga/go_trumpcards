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

// suitStr returns the English suit name for a design constant. Shared across
// categories (Belote/BigTwo are classic; Euchre/FiveHundred/Schnapsen are
// solo), so it lives in this untagged core file (#2126).
func suitStr(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "Spade"
	case CardDesignClover:
		return "Club"
	case CardDesignHeart:
		return "Heart"
	case CardDesignDiamond:
		return "Diamond"
	}
	return "Unknown"
}
