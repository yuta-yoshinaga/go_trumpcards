package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

const (
	daifugoRankMin = 1
	daifugoRankMax = 4
)

// daifugoRankName は大富豪のプレイヤーランク（1-4）に対応する日本語名を返します。
// 範囲外のランクには「不明」を返します。
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

// daifugoSuitName は大富豪のスート縛りに使用するスート名の文字列を返します。
// 4つの標準スート以外には空文字列 "" を返します。
func daifugoSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond:
		return cardDesignToString(suit)
	default:
		return ""
	}
}
