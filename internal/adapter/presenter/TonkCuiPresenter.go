package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// tonkPlayerStr returns the display string for a single Tonk player.
func tonkPlayerStr(player *domain.TonkPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("tonk.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// TonkCuiPresenter renders the Tonk CUI view.
type TonkCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *TonkCuiPresenter) Output(g interfaces.TonkGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tonk.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tonk.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("tonk.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(tonkPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("tonk.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.TonkPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tonk.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("tonk.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("tonk.promptDrawHelpDiscard") + "\n")
		case domain.TonkPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tonk.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("tonk.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("tonk.promptKnockHelp") + "\n")
		case domain.TonkPhaseRoundEnd:
			if g.GetIsTonk() {
				b.WriteString(i18n.T("tonk.promptDealtTonk") + "\n")
			}
			b.WriteString(i18n.T("tonk.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("tonk.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TonkCuiPresenter) ActionLogOutput(g interfaces.TonkGame) string {
	return actionLogOutputText(g)
}
