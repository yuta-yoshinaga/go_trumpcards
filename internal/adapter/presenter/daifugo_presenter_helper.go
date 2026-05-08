package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

const (
	daifugoRankMin = 1
	daifugoRankMax = 4
)

// daifugoRankName returns the localized Daifugo rank label (1=Daifugo,
// 2=Fugo, 3=Heimin, 4=Daihinmin). Out-of-range ranks fall back to the
// generic "unknown" key.
func daifugoRankName(rank int) string {
	switch rank {
	case 1:
		return i18n.T("daifugo.rank1")
	case 2:
		return i18n.T("daifugo.rank2")
	case 3:
		return i18n.T("daifugo.rank3")
	case 4:
		return i18n.T("daifugo.rank4")
	default:
		return i18n.T("daifugo.rankUnknown")
	}
}

// daifugoSuitName returns the suit name for the locked-suit banner, or
// empty string for unrecognized suit values.
func daifugoSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond:
		return cardDesignToString(suit)
	default:
		return ""
	}
}
