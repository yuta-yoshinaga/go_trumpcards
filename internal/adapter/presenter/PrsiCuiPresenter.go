package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// prsiPlayerStr returns the display string for a single Prsi player.
func prsiPlayerStr(player *domain.PrsiPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("prsi.playerLine",
		"name", cuiPlayerName(player, i),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// PrsiCuiPresenter renders the Prší CUI view.
type PrsiCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PrsiCuiPresenter) Output(g interfaces.PrsiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("prsi.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("prsi.header",
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		// Top of discard pile
		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("prsi.discardLine", "card", cuiCardStr(top)))
			if g.GetPenaltyDrawCount() > 0 {
				b.WriteString(i18n.Tf("prsi.penaltyDraw",
					"count", strconv.Itoa(g.GetPenaltyDrawCount())))
			}
			b.WriteString("\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(prsiPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("prsi.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		if g.GetPhase() == domain.PrsiPhasePlay {
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("prsi.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("prsi.promptPlayHelp") + "\n")
			b.WriteString(i18n.T("prsi.promptDrawHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PrsiCuiPresenter) ActionLogOutput(g interfaces.PrsiGame) string {
	return actionLogOutputText(g)
}
