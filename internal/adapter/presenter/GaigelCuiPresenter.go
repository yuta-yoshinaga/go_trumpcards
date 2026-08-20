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

// gaigelPlayerStr returns the display string for a single Gaigel player.
func gaigelPlayerStr(player *domain.GaigelPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("gaigel.playerLine",
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

// GaigelCuiPresenter renders the Gaigel CUI view.
type GaigelCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *GaigelCuiPresenter) Output(g interfaces.GaigelGame, lastErr error) string {
	return buildCuiOutput(i18n.T("gaigel.helpTitle"), func(out *strings.Builder) {
		out.WriteString(i18n.Tf("gaigel.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")
		dealerIdx := g.GetDealerIdx()
		out.WriteString(i18n.Tf("gaigel.dealer",
			"name", cuiPlayerName(g.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		if trumpSuit := g.GetTrumpSuit(); trumpSuit > 0 {
			// **表向きの1枚が切り札を決め、それが最後に山から引かれる札** (#5686)。
			// Web はその実カードを出しているのに、CUI はスートしか出していなかった。
			// 山が尽きるとこの札は引かれて無くなるので、そのときはスートだけ。
			if card := g.GetTrumpCard(); card != nil {
				out.WriteString(i18n.Tf("gaigel.trumpLineWithCard",
					"suit", cuiSuitName(trumpSuit),
					"card", cuiCardStr(card),
					"stock", strconv.Itoa(g.GetStockRemaining())) + "\n")
			} else {
				out.WriteString(i18n.Tf("gaigel.trumpLine",
					"suit", cuiSuitName(trumpSuit),
					"stock", strconv.Itoa(g.GetStockRemaining())) + "\n")
			}
		} else {
			out.WriteString(i18n.T("gaigel.trumpUndecided") + "\n")
		}

		out.WriteString(i18n.Tf("gaigel.teamScoreLine",
			"t0", strconv.Itoa(g.GetTeamScore(0)),
			"t1", strconv.Itoa(g.GetTeamScore(1))) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			out.WriteString(gaigelPlayerStr(g.GetPlayer(i), i))
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
			banner := i18n.Tf("gaigel.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			out.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.GaigelPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			out.WriteString(i18n.Tf("gaigel.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if idxs := g.GetMarriageIndices(currentIdx); len(idxs) > 0 {
				out.WriteString(i18n.T("gaigel.promptMarriageHint") + "\n")
				// On the human's turn, name the K/Q cards (and their hand indices)
				// that can be declared; never for a CPU, to avoid leaking its hand.
				if human := g.GetPlayer(currentIdx); human != nil && human.GetIsHuman() {
					cards := make([]string, len(idxs))
					for i, idx := range idxs {
						cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(human.GetCard(idx))
					}
					out.WriteString(i18n.Tf("gaigel.promptMarriageCards",
						"cards", strings.Join(cards, ", ")) + "\n")
				}
			}
			out.WriteString(i18n.T("gaigel.promptPlayHelp") + "\n")
		case domain.GaigelPhaseTrickEnd:
			out.WriteString(i18n.T("gaigel.promptTrickEnd") + "\n")
			out.WriteString(i18n.T("gaigel.promptTrickEndHelp") + "\n")
		case domain.GaigelPhaseRoundEnd:
			out.WriteString(i18n.T("gaigel.promptRoundEnd") + "\n")
			out.WriteString(i18n.T("gaigel.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Gaigel hint.
func (p *GaigelCuiPresenter) HintOutput(g interfaces.GaigelGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("gaigel.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, gaigelHintReasonKeys)
	if hint.CardIndex == nil {
		return i18n.T("gaigel.hintNone") + "\n"
	}
	player := g.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	if hint.IsMarriage {
		return color.Yellow(i18n.Tf("gaigel.hintMarriage",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("gaigel.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GaigelCuiPresenter) ActionLogOutput(g interfaces.GaigelGame) string {
	return actionLogOutputTextForSeats[*domain.GaigelPlayer](g)
}

// gaigelHintReasonKeys maps Gaigel-specific hint-reason identifiers to their
// i18n keys, consumed by hintReasonStr.
var gaigelHintReasonKeys = map[string]string{
	"lead_trump":  "gaigel.hintReasonLeadTrump",
	"lead_low":    "gaigel.hintReasonLeadLow",
	"lead_value":  "gaigel.hintReasonLeadValue",
	"follow_cut":  "gaigel.hintReasonFollowCut",
	"follow_win":  "gaigel.hintReasonFollowWin",
	"follow_dump": "gaigel.hintReasonFollowDump",
	"marriage":    "gaigel.hintReasonMarriage",
}
