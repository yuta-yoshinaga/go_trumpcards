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

// suitKeyOf returns the i18n key for a suit design.
func suitKeyOf(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "common.suit.spade"
	case CardDesignClover:
		return "common.suit.club"
	case CardDesignHeart:
		return "common.suit.heart"
	case CardDesignDiamond:
		return "common.suit.diamond"
	default:
		return "common.suit.unknown"
	}
}

// trumpKeyOf は切り札スートの i18n キーを返す。suitKeyOf と違い、
// スート以外の値は「切り札なし」を意味するゲームで使うため
// common.suit.notrump に落とす。
func trumpKeyOf(suit int) string {
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return "common.suit.notrump"
	}
	return suitKeyOf(suit)
}

// cardStr returns the display string for a card (suit glyph + rank). Shared
// across categories — over a hundred game files render a card this way — so it
// lives here rather than in any one game's file.
func cardStr(card *Card) string {
	suits := map[int]string{
		CardDesignSpade:   "♠",
		CardDesignClover:  "♣",
		CardDesignHeart:   "♥",
		CardDesignDiamond: "♦",
	}
	values := map[int]string{
		1: "A", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7",
		8: "8", 9: "9", 10: "10", 11: "J", 12: "Q", 13: "K",
	}
	s, ok := suits[card.GetDesign()]
	if !ok {
		s = "?"
	}
	v, ok := values[card.GetValue()]
	if !ok {
		v = "?"
	}
	return s + v
}
