package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// isRedSuit returns true for heart and diamond suits.
func isRedSuit(design int) bool {
	return design == domain.CardDesignHeart || design == domain.CardDesignDiamond
}

// cuiCardList is the minimal type constraint required by formatCardList.
type cuiCardList interface {
	GetCardsSize() int
	GetCard(idx int) *domain.Card
}

// cardFormatter formats a single card into a string.
type cardFormatter func(card *domain.Card) string

// formatCardList formats all cards in a cuiCardList using the given formatter and separator.
// When indexed is true, each card is prefixed with "[N]".
func formatCardList(hand cuiCardList, fmtCard cardFormatter, sep string, indexed bool) string {
	parts := make([]string, hand.GetCardsSize())
	for i := range parts {
		s := fmtCard(hand.GetCard(i))
		if indexed {
			s = fmt.Sprintf("[%d]%s", i, s)
		}
		parts[i] = s
	}
	return strings.Join(parts, sep)
}

// formatCardSlice formats a card slice using the given formatter and separator.
func formatCardSlice(cards []*domain.Card, fmtCard cardFormatter, sep string) string {
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = fmtCard(c)
	}
	return strings.Join(parts, sep)
}

// cuiCardListStr returns a comma-separated card string for all cards in hand.
func cuiCardListStr(hand cuiCardList) string {
	return formatCardList(hand, cuiCardStr, ",", false)
}

// cuiPlayer is the minimal type constraint required by cuiPlayerName.
type cuiPlayer interface {
	comparable
	GetIsHuman() bool
}

// suitNames maps design constants to suit name strings.
// Index 0 is unused (joker); indices 1–4 correspond to CardDesignSpade–CardDesignDiamond.
var suitNames = []string{"", "SPADE", "CLOVER", "HEART", "DIAMOND"}

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

// cuiRankLabel returns the card-face label for a rank value: A for 1, J/Q/K for
// 11/12/13, and the plain number otherwise (7–10). Locale-independent — matches
// the card-face notation used elsewhere in the UI. Used by the Watten CUI
// presenter for Schlag-rank display.
func cuiRankLabel(rank int) string {
	switch rank {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	default:
		return strconv.Itoa(rank)
	}
}

// cuiPokerHandName returns the localized display name for a poker hand rank
// (0=High Card .. 10=Five of a Kind), resolved via the shared pokerHandRank*
// keys in cui_common. Out-of-range ranks fall back to the raw English
// domain.PokerHandNames entry (or "" when the index is invalid). Used by the
// UltimateTexasHoldem and MississippiStud CUI presenters.
func cuiPokerHandName(rank int) string {
	if rank < 0 || rank >= len(domain.PokerHandNames) {
		return ""
	}
	return i18n.T("pokerHandRank" + strconv.Itoa(rank))
}

// cuiPlayerName returns the human-friendly display name for a player:
// "You" / "あなた" for the human, "CPU N" for CPU opponents, or
// "UNKNOWN" if the player is nil/zero. Locale-aware via i18n.T (issue
// #1699 Phase 1). Used by OldMaid, Daifugo, Sevens, Doubt, Poker,
// Holdem, Omaha, Hearts, Spades, CrazyEights, GinRummy, and Memory.
func cuiPlayerName[P cuiPlayer](player P, idx int) string {
	var zero P
	if player == zero {
		return i18n.T("cuiPlayerUnknown")
	}
	if player.GetIsHuman() {
		return color.Bold(i18n.T("cuiPlayerYou"))
	}
	return color.Bold(i18n.Tf("cuiPlayerCpu", "idx", strconv.Itoa(idx)))
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

// cuiIndexedCardListStr returns a double-space separated indexed card string.
// e.g. "[0]SPADE 5  [1]HEART 3"
func cuiIndexedCardListStr(hand cuiCardList) string {
	return formatCardList(hand, cuiCardStr, "  ", true)
}

// cuiCardListStrEmoji returns a double-space separated emoji card string (no index).
// e.g. "♠5  ♥3"
func cuiCardListStrEmoji(hand cuiCardList) string {
	return formatCardList(hand, cuiCardStrEmoji, "  ", false)
}

// cuiIndexedCardListStrEmoji returns a double-space separated indexed emoji card string.
// e.g. "[0]♠5  [1]♥3"
func cuiIndexedCardListStrEmoji(hand cuiCardList) string {
	return formatCardList(hand, cuiCardStrEmoji, "  ", true)
}

// cuiCardSliceStr returns a comma-space separated card string from a card slice.
// e.g. "SPADE 5, HEART 3"
func cuiCardSliceStr(cards []*domain.Card) string {
	return formatCardSlice(cards, cuiCardStr, ", ")
}

// cuiCardSliceStrEmoji returns a double-space separated emoji card string from a card slice.
// e.g. "♠5  ♥3"
func cuiCardSliceStrEmoji(cards []*domain.Card) string {
	return formatCardSlice(cards, cuiCardStrEmoji, "  ")
}

// cuiCaptureHintLine annotates each hand card with the table cards it can
// capture, e.g. `[0]♠5 → 場[1][3]`.
//
// Shared by the fishing games (Basra, Tablanet), which all get the pairing from
// the domain's own `GetCaptureOptions`. **Recomputing it per presenter is how
// the note starts disagreeing with what the server will accept** (#4922).
// Cards that capture nothing get no note; an empty result means no line at all.
func cuiCaptureHintLine(hand cuiCardList, opts map[int][]int, key string) string {
	if len(opts) == 0 {
		return ""
	}
	notes := make([]string, 0, hand.GetCardsSize())
	for h := range hand.GetCardsSize() {
		tableIdxs := opts[h]
		if len(tableIdxs) == 0 {
			continue
		}
		marks := make([]string, len(tableIdxs))
		for j, ti := range tableIdxs {
			marks[j] = "[" + strconv.Itoa(ti) + "]"
		}
		notes = append(notes, i18n.Tf(key,
			"hand", "["+strconv.Itoa(h)+"]"+cuiCardStr(hand.GetCard(h)),
			"table", strings.Join(marks, "")))
	}
	if len(notes) == 0 {
		return ""
	}
	return strings.Join(notes, "  ")
}
