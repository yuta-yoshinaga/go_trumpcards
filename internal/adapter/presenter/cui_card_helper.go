package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// isRedSuit returns true for heart and diamond suits.
func isRedSuit(design int) bool {
	return design == domain.CardDesignHeart || design == domain.CardDesignDiamond
}

// cuiCardList is the minimal type constraint required by cuiCardListStr.
type cuiCardList interface {
	GetCardsSize() int
	GetCard(idx int) *domain.Card
}

// cuiCardListStr returns a comma-separated card string for all cards in hand.
func cuiCardListStr(hand cuiCardList) string {
	parts := make([]string, hand.GetCardsSize())
	for i := range parts {
		parts[i] = cuiCardStr(hand.GetCard(i))
	}
	return strings.Join(parts, ",")
}

// cuiPlayer is the minimal type constraint required by cuiPlayerName.
type cuiPlayer interface {
	comparable
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
	s := name + " " + strconv.Itoa(card.GetValue())
	if isRedSuit(card.GetDesign()) {
		return color.Red(s)
	}
	return s
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
	s := fmt.Sprintf("%s%d", designs[d], card.GetValue())
	if isRedSuit(card.GetDesign()) {
		return color.Red(s)
	}
	return s
}

// cuiSuitName returns the suit name string for a given design constant.
// Used by Daifugo and Sevens CUI presenters.
func cuiSuitName(suit int) string {
	if suit > 0 && suit < len(suitNames) {
		return suitNames[suit]
	}
	return "UNKNOWN"
}

// cuiPlayerName returns "あなた" for human players, "CPU N" for CPU players,
// or "UNKNOWN" if the player is nil/zero.
// Used by OldMaid, Daifugo, Sevens, Doubt, Poker, Holdem, Omaha, Hearts,
// Spades, CrazyEights, GinRummy, and Memory CUI presenters.
func cuiPlayerName[P cuiPlayer](player P, idx int) string {
	var zero P
	if player == zero {
		return "UNKNOWN"
	}
	if player.GetIsHuman() {
		return color.Bold("あなた")
	}
	return color.Bold(fmt.Sprintf("CPU %d", idx))
}

// cuiPlayerWithStyle is the type constraint for players that have a play style.
type cuiPlayerWithStyle interface {
	cuiPlayer
	GetPlayStyleName() string
}

// cuiPlayerNameWithStyle returns cuiPlayerName with play style suffix for CPU.
// Used by Poker, Holdem, and Omaha CUI presenters.
func cuiPlayerNameWithStyle[P cuiPlayerWithStyle](player P, idx int) string {
	name := cuiPlayerName(player, idx)
	if !player.GetIsHuman() {
		name = fmt.Sprintf("%s (%s)", name, player.GetPlayStyleName())
	}
	return name
}

// cuiBettingActionName returns the Japanese action name for betting actions.
// Used by Poker and Holdem CUI presenters.
func cuiBettingActionName(action int) string {
	if name, ok := bettingActionNames[action]; ok {
		return name
	}
	return "不明"
}

// cuiIndexedCardListStr returns a double-space separated indexed card string.
// e.g. "[0]SPADE 5  [1]HEART 3"
func cuiIndexedCardListStr(hand cuiCardList) string {
	parts := make([]string, hand.GetCardsSize())
	for i := range parts {
		parts[i] = fmt.Sprintf("[%d]%s", i, cuiCardStr(hand.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// cuiCardListStrEmoji returns a double-space separated emoji card string (no index).
// e.g. "♠5  ♥3"
func cuiCardListStrEmoji(hand cuiCardList) string {
	parts := make([]string, hand.GetCardsSize())
	for i := range parts {
		parts[i] = cuiCardStrEmoji(hand.GetCard(i))
	}
	return strings.Join(parts, "  ")
}

// cuiIndexedCardListStrEmoji returns a double-space separated indexed emoji card string.
// e.g. "[0]♠5  [1]♥3"
func cuiIndexedCardListStrEmoji(hand cuiCardList) string {
	parts := make([]string, hand.GetCardsSize())
	for i := range parts {
		parts[i] = fmt.Sprintf("[%d]%s", i, cuiCardStrEmoji(hand.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// cuiCardSliceStr returns a comma-space separated card string from a card slice.
// e.g. "SPADE 5, HEART 3"
func cuiCardSliceStr(cards []*domain.Card) string {
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = cuiCardStr(c)
	}
	return strings.Join(parts, ", ")
}

// cuiCardSliceStrEmoji returns a double-space separated emoji card string from a card slice.
// e.g. "♠5  ♥3"
func cuiCardSliceStrEmoji(cards []*domain.Card) string {
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = cuiCardStrEmoji(c)
	}
	return strings.Join(parts, "  ")
}
