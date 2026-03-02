package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// cuiPlayer is the minimal interface required by cuiPlayerName.
type cuiPlayer interface {
	GetIsHuman() bool
}

// cuiCardStr returns a text-based card string (e.g. "SPADE 5", "JOKER", "??").
// Used by BlackJack, OldMaid, Daifugo, Sevens, and Doubt CUI presenters.
func cuiCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	switch card.GetDesign() {
	case domain.CardDesignJoker:
		return "JOKER"
	case domain.CardDesignSpade:
		return "SPADE " + strconv.Itoa(card.GetValue())
	case domain.CardDesignClover:
		return "CLOVER " + strconv.Itoa(card.GetValue())
	case domain.CardDesignHeart:
		return "HEART " + strconv.Itoa(card.GetValue())
	case domain.CardDesignDiamond:
		return "DIAMOND " + strconv.Itoa(card.GetValue())
	default:
		return "UNKNOWN"
	}
}

// cuiCardStrEmoji returns an emoji-based card string (e.g. "♠5", "🃏0").
// Used by Poker and Holdem CUI presenters.
func cuiCardStrEmoji(card *domain.Card) string {
	designs := []string{"🃏", "♠", "♣", "♥", "♦"}
	d := card.GetDesign()
	if d < 0 || d >= len(designs) {
		d = 0
	}
	return fmt.Sprintf("%s%d", designs[d], card.GetValue())
}

// cuiSuitName returns the suit name string for a given design constant.
// Used by Daifugo and Sevens CUI presenters.
func cuiSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "SPADE"
	case domain.CardDesignClover:
		return "CLOVER"
	case domain.CardDesignHeart:
		return "HEART"
	case domain.CardDesignDiamond:
		return "DIAMOND"
	default:
		return "UNKNOWN"
	}
}

// cuiPlayerName returns "あなた" for human players, "CPU N" for CPU players.
// Used by OldMaid, Daifugo, Sevens, and Doubt CUI presenters.
func cuiPlayerName(player cuiPlayer, idx int) string {
	if player.GetIsHuman() {
		return "あなた"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// cuiBettingActionName returns the Japanese action name for betting actions.
// Used by Poker and Holdem CUI presenters.
func cuiBettingActionName(action int) string {
	switch action {
	case domain.PokerActionFold:
		return "フォールド"
	case domain.PokerActionCheck:
		return "チェック"
	case domain.PokerActionCall:
		return "コール"
	case domain.PokerActionBet:
		return "ベット"
	case domain.PokerActionRaise:
		return "レイズ"
	case domain.PokerActionAllIn:
		return "オールイン"
	default:
		return "不明"
	}
}
