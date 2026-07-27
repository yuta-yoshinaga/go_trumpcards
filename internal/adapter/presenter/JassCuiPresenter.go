//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// jassPlayerStr returns the display string for a single Jass player.
func jassPlayerStr(player *domain.JassPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("jass.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// JassCuiPresenter renders the Jass CUI view.
type JassCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *JassCuiPresenter) Output(g interfaces.JassGame, lastErr error) string {
	return buildCuiOutput(i18n.T("jass.helpTitle"), func(out *strings.Builder) {
		out.WriteString(i18n.Tf("jass.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")
		dealerIdx := g.GetDealerIdx()
		out.WriteString(i18n.Tf("jass.dealer",
			"name", cuiPlayerName(g.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		if trumpSuit := g.GetTrumpSuit(); trumpSuit > 0 {
			out.WriteString(i18n.Tf("jass.trumpLine",
				"suit", cuiSuitName(trumpSuit),
				"team", strconv.Itoa(g.GetMakerTeam())) + "\n")
		} else {
			out.WriteString(i18n.T("jass.trumpUndecided") + "\n")
		}

		out.WriteString(i18n.Tf("jass.teamScoreLine",
			"t0", strconv.Itoa(g.GetTeamScore(0)),
			"t1", strconv.Itoa(g.GetTeamScore(1))) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			out.WriteString(jassPlayerStr(g.GetPlayer(i), i))
		}

		out.WriteString("----------\n")

		trick := g.GetCurrentTrick()
		cuiTrickBlock(out, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(out, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("jass.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			out.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.JassPhaseBidTrump:
			bidIdx := g.GetBidPlayerIdx()
			out.WriteString(i18n.Tf("jass.promptBidTrump",
				"name", cuiPlayerName(g.GetPlayer(bidIdx), bidIdx)) + "\n")
			out.WriteString(i18n.T("jass.promptBidTrumpHelp") + "\n")
		case domain.JassPhaseBidPartner:
			bidIdx := g.GetBidPlayerIdx()
			out.WriteString(i18n.Tf("jass.promptBidPartner",
				"name", cuiPlayerName(g.GetPlayer(bidIdx), bidIdx)) + "\n")
			out.WriteString(i18n.T("jass.promptBidPartnerHelp") + "\n")
		case domain.JassPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			out.WriteString(i18n.Tf("jass.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			out.WriteString(i18n.T("jass.promptPlayHelp") + "\n")
		case domain.JassPhaseTrickEnd:
			out.WriteString(i18n.T("jass.promptTrickEnd") + "\n")
			out.WriteString(i18n.T("jass.promptTrickEndHelp") + "\n")
		case domain.JassPhaseRoundEnd:
			out.WriteString(i18n.T("jass.promptRoundEnd") + "\n")
			out.WriteString(i18n.T("jass.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Jass hint.
func (p *JassCuiPresenter) HintOutput(g interfaces.JassGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("jass.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, nil)
	if hint.Schieben != nil && *hint.Schieben {
		return color.Yellow(i18n.Tf("jass.hintSchieben", "reason", reason)) + "\n"
	}
	if hint.Suit != nil {
		return color.Yellow(i18n.Tf("jass.hintTrump",
			"suit", cuiSuitName(*hint.Suit),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("jass.hintNone") + "\n"
	}
	player := g.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("jass.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *JassCuiPresenter) ActionLogOutput(g interfaces.JassGame) string {
	return actionLogOutputText(g)
}
