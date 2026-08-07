//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// napoleonSuitGlyphs maps the trump-suit constant to a Unicode glyph used
// in the trump banner and the bid hint. The glyphs are language-neutral
// so they live next to the presenter rather than in the locale files.
var napoleonSuitGlyphs = map[int]string{
	domain.CardDesignSpade:   "♠",
	domain.CardDesignClover:  "♣",
	domain.CardDesignHeart:   "♥",
	domain.CardDesignDiamond: "♦",
}

// napoleonBidStr renders the bid column for the player line.
func napoleonBidStr(bid int) string {
	switch {
	case bid == 0:
		return i18n.T("napoleon.bidPass")
	case bid > 0:
		return strconv.Itoa(bid)
	default:
		return i18n.T("napoleon.bidPending")
	}
}

// napoleonRoleStr returns the role badge (Napoleon / adjutant) prefixed
// with a leading space, or "" when the player has no badge to display.
func napoleonRoleStr(player *domain.NapoleonPlayer, adjRevealed bool) string {
	if player.GetIsNapoleon() {
		return " " + i18n.T("napoleon.roleNapoleon")
	}
	if adjRevealed && player.GetIsAdjutant() {
		return " " + i18n.T("napoleon.roleAdjutant")
	}
	return ""
}

// napoleonPlayerStr returns the display string for a single Napoleon player.
func napoleonPlayerStr(player *domain.NapoleonPlayer, i int, adjRevealed bool) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("napoleon.playerLine",
		"name", cuiPlayerName(player, i),
		"role", napoleonRoleStr(player, adjRevealed),
		"bid", napoleonBidStr(player.GetBid()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"pictures", strconv.Itoa(player.GetPictureCards()),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// NapoleonCuiPresenter renders the Napoleon CUI view.
type NapoleonCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *NapoleonCuiPresenter) Output(n interfaces.NapoleonGame, lastErr error) string {
	return buildCuiOutput(i18n.T("napoleon.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("napoleon.round",
			"round", strconv.Itoa(n.GetRoundNumber()),
			"trick", strconv.Itoa(n.GetTrickNumber())))
		b.WriteString("\n")

		if n.GetTrumpSuit() > 0 {
			b.WriteString(i18n.Tf("napoleon.trump", "suit", napoleonSuitGlyphs[n.GetTrumpSuit()]))
			b.WriteString("\n")
		}

		if adjCard := n.GetAdjutantCard(); adjCard != nil {
			b.WriteString(i18n.Tf("napoleon.adjutantCard", "card", napoleonCuiCardStr(adjCard)))
			if n.GetAdjutantRevealed() {
				b.WriteString(i18n.T("napoleon.adjutantRevealed"))
			} else {
				b.WriteString(i18n.T("napoleon.adjutantHidden"))
			}
			b.WriteString("\n")
		}

		if n.GetHighestBid() > 0 {
			b.WriteString(i18n.Tf("napoleon.highestBid", "bid", strconv.Itoa(n.GetHighestBid())))
			b.WriteString("\n")
		}

		// Napoleon-army face-card progress vs the bid target. The adjutant's
		// captures are only added once revealed, to avoid leaking their identity
		// (mirrors the web napoleonFaceProgress logic).
		if n.GetPhase() == domain.NapoleonPhasePlay {
			napIdx := n.GetNapoleonIdx()
			if napIdx >= 0 {
				collected := n.GetPlayer(napIdx).GetPictureCards()
				if n.GetAdjutantRevealed() {
					if adjIdx := n.GetAdjutantIdx(); adjIdx >= 0 && adjIdx != napIdx {
						collected += n.GetPlayer(adjIdx).GetPictureCards()
					}
				}
				bid := n.GetHighestBid()
				line := i18n.Tf("napoleon.armyProgress",
					"collected", strconv.Itoa(collected),
					"bid", strconv.Itoa(bid))
				if bid > 0 && collected >= bid {
					line = color.Green(line)
				}
				b.WriteString(line + "\n")
			}
		}

		// Player rows
		for i := 0; i < n.GetPlayerCnt(); i++ {
			b.WriteString(napoleonPlayerStr(n.GetPlayer(i), i, n.GetAdjutantRevealed()))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := n.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.NapoleonTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.NapoleonTrickCard) string { return napoleonCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(n.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if n.GetGameEndFlag() {
			if n.GetWinnerTeam() == domain.NapoleonWinnerNapoleon {
				b.WriteString(color.Green(i18n.T("napoleon.gameEndNapoleonWin")) + "\n")
			} else {
				b.WriteString(color.Green(i18n.T("napoleon.gameEndAlliedWin")) + "\n")
			}
			return
		}
		switch n.GetPhase() {
		case domain.NapoleonPhaseBid:
			bidIdx := n.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("napoleon.promptBid",
				"name", cuiPlayerName(n.GetPlayer(bidIdx), bidIdx)))
			b.WriteString("\n")
			b.WriteString(i18n.T("napoleon.promptBidHelp") + "\n")
		case domain.NapoleonPhaseTrumpDeclaration:
			b.WriteString(i18n.T("napoleon.promptTrumpHeader") + "\n")
			b.WriteString(i18n.T("napoleon.promptTrumpDeclareHelp") + "\n")
			b.WriteString(i18n.T("napoleon.promptTrumpSuitLegend") + "\n")
		case domain.NapoleonPhaseKittyExchange:
			b.WriteString(i18n.T("napoleon.promptKittyHeader") + "\n")
			b.WriteString(i18n.T("napoleon.promptKittyHelp") + "\n")
		case domain.NapoleonPhasePlay:
			currentIdx := n.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("napoleon.promptPlayTurn",
				"name", cuiPlayerName(n.GetPlayer(currentIdx), currentIdx)))
			b.WriteString("\n")
			b.WriteString(i18n.T("napoleon.promptPlayHelp") + "\n")
		case domain.NapoleonPhaseTrickEnd:
			b.WriteString(i18n.T("napoleon.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("napoleon.promptTrickEndHelp") + "\n")
		case domain.NapoleonPhaseRoundEnd:
			b.WriteString(i18n.T("napoleon.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("napoleon.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Napoleon hint.
func (p *NapoleonCuiPresenter) HintOutput(n interfaces.NapoleonGame) string {
	hint := n.GetHint()
	if hint == nil {
		return i18n.T("napoleon.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, napoleonHintReasonKeys)
	switch {
	case hint.Bid != nil:
		return color.Yellow(i18n.Tf("napoleon.hintBid",
			"bid", strconv.Itoa(*hint.Bid),
			"reason", reason)) + "\n"
	case hint.TrumpSuit != nil:
		return color.Yellow(i18n.Tf("napoleon.hintTrump",
			"suit", napoleonSuitGlyphs[*hint.TrumpSuit],
			"reason", reason)) + "\n"
	case hint.DiscardIndex != nil:
		// **人間が席 0 とは限らない。**ヒントの数値自体はドメインが
		// findHumanIdx で解決しているのに、表示だけが決め打ちだった (#4689)。
		card := n.GetPlayer(n.GetHumanIdx()).GetCard(*hint.DiscardIndex)
		return color.Yellow(i18n.Tf("napoleon.hintDiscard",
			"idx", strconv.Itoa(*hint.DiscardIndex),
			"card", napoleonCuiCardStr(card),
			"reason", reason)) + "\n"
	case hint.CardIndex != nil:
		card := n.GetPlayer(n.GetHumanIdx()).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("napoleon.hintCard",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", napoleonCuiCardStr(card),
			"reason", reason)) + "\n"
	}
	return i18n.T("napoleon.hintNone") + "\n"
}

// napoleonCuiCardStr renders a card with Napoleon-flavoured rules
// (Joker → "Joker"); other cards use the standard cuiCardStr formatter.
func napoleonCuiCardStr(card *domain.Card) string {
	if card.GetDesign() == domain.CardDesignJoker {
		return "Joker"
	}
	return cuiCardStr(card)
}

// napoleonHintReasonKeys maps a hint-reason identifier specific to Napoleon
// to its i18n key. The displayed string follows the active locale via
// i18n.T (issue #1699 Phase 2).
var napoleonHintReasonKeys = map[string]string{
	"strategic_declare": "napoleon.hintReasonStrategicDeclare",
	"strategic_discard": "napoleon.hintReasonStrategicDiscard",
	"play_joker":        "napoleon.hintReasonPlayJoker",
	"discard_low":       "napoleon.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *NapoleonCuiPresenter) ActionLogOutput(n interfaces.NapoleonGame) string {
	return actionLogOutputText(n)
}
