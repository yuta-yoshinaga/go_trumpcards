package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// daifugoRankName returns the Japanese name for a Daifugo player rank (1–4).
// Returns "不明" for ranks outside the valid range.
func daifugoRankName(rank int) string {
	switch rank {
	case 1:
		return "大富豪"
	case 2:
		return "富豪"
	case 3:
		return "平民"
	case 4:
		return "大貧民"
	default:
		return "不明"
	}
}

// daifugoSuitName returns the suit name string for a Daifugo suit restriction.
// Returns "" for suits other than the four standard suits.
func daifugoSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond:
		return cardDesignToString(suit)
	default:
		return ""
	}
}
