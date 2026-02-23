package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// cardDesignToString カードデザイン定数を文字列に変換する共通ヘルパー関数
func cardDesignToString(design int) string {
	switch design {
	case domain.CardDesignSpade:
		return "SPADE"
	case domain.CardDesignClover:
		return "CLOVER"
	case domain.CardDesignHeart:
		return "HEART"
	case domain.CardDesignDiamond:
		return "DIAMOND"
	default:
		return "JOKER"
	}
}
