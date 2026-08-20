package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// catchTenPlayerStr returns the display string for a single Catch the Ten player.
func catchTenPlayerStr(player *domain.CatchTenPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("catchten.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// CatchTenCuiPresenter renders the Catch the Ten CUI view.
type CatchTenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CatchTenCuiPresenter) Output(g interfaces.CatchTenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("catchten.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("catchten.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")
		b.WriteString(i18n.Tf("catchten.trumpLine",
			"suit", suitDisplayName(g.GetTrumpSuit())) + "\n")
		b.WriteString(i18n.Tf("catchten.teamScoreLine",
			"t0", strconv.Itoa(g.GetTeamScore(0)),
			"t1", strconv.Itoa(g.GetTeamScore(1))) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(catchTenPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		trick := g.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			var banner string
			if g.GetWinnerTeam() == domain.CatchTenDrawTeam {
				banner = i18n.T("catchten.gameEndDraw")
			} else {
				banner = i18n.Tf("catchten.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			}
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.CatchTenPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("catchten.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("catchten.promptPlay") + "\n")
		case domain.CatchTenPhaseTrickEnd:
			b.WriteString(i18n.T("catchten.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("catchten.promptTrickEndHelp") + "\n")
		case domain.CatchTenPhaseRoundEnd:
			b.WriteString(i18n.T("catchten.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("catchten.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Catch the Ten hint.
func (p *CatchTenCuiPresenter) HintOutput(g interfaces.CatchTenGame) string {
	hint := g.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("catchten.hintNone") + "\n"
	}
	player := g.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("catchten.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, catchTenHintReasonKeys))) + "\n"
}

// catchTenHintReasonKeys maps Catch the Ten-specific hint-reason identifiers
// to their i18n keys. Reasons not listed here fall through to cui_common.
var catchTenHintReasonKeys = map[string]string{
	"trump_cut": "catchten.hintReasonTrumpCut",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CatchTenCuiPresenter) ActionLogOutput(g interfaces.CatchTenGame) string {
	return actionLogOutputTextForSeats[*domain.CatchTenPlayer](g)
}
