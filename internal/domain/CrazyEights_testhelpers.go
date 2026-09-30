//go:build test

package domain

// suitName は Crazy Eights と Tarneeb の内部テストで使う表示名。
func suitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "♠"
	case CardDesignClover:
		return "♣"
	case CardDesignHeart:
		return "♥"
	case CardDesignDiamond:
		return "♦"
	default:
		return "?"
	}
}
