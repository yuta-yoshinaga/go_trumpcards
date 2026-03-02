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

// suitNames maps design constants to suit name strings.
// Index 0 is unused (joker); indices 1–4 correspond to CardDesignSpade–CardDesignDiamond.
var suitNames = []string{"", "SPADE", "CLOVER", "HEART", "DIAMOND"}

// bettingActionNames maps betting action constants to Japanese action name strings.
var bettingActionNames = map[int]string{
	domain.PokerActionFold:  "フォールド",
	domain.PokerActionCheck: "チェック",
	domain.PokerActionCall:  "コール",
	domain.PokerActionBet:   "ベット",
	domain.PokerActionRaise: "レイズ",
	domain.PokerActionAllIn: "オールイン",
}

// cuiCardStr returns a text-based card string (e.g. "SPADE 5", "JOKER", "??").
// Used by BlackJack, OldMaid, Daifugo, Sevens, and Doubt CUI presenters.
func cuiCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	if card.GetDesign() == domain.CardDesignJoker {
		return "JOKER"
	}
	name := cuiSuitName(card.GetDesign())
	if name == "UNKNOWN" {
		return "UNKNOWN"
	}
	return name + " " + strconv.Itoa(card.GetValue())
}

// cuiCardStrEmoji returns an emoji-based card string (e.g. "♠5", "🃏0").
// Used by Poker and Holdem CUI presenters.
func cuiCardStrEmoji(card *domain.Card) string {
	if card == nil {
		return "??"
	}
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
	if suit > 0 && suit < len(suitNames) {
		return suitNames[suit]
	}
	return "UNKNOWN"
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
	if name, ok := bettingActionNames[action]; ok {
		return name
	}
	return "不明"
}
