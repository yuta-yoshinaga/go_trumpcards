package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ginRummyPlayerStr returns the display string for a single GinRummy player.
func ginRummyPlayerStr(player *domain.GinRummyPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("ginrummy.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// GinRummyCuiPresenter renders the Gin Rummy CUI view.
type GinRummyCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *GinRummyCuiPresenter) Output(g interfaces.GinRummyGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ginrummy.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("ginrummy.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("ginrummy.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(ginRummyPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("ginrummy.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.GinRummyPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("ginrummy.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("ginrummy.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("ginrummy.promptDrawHelpDiscard") + "\n")
		case domain.GinRummyPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("ginrummy.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("ginrummy.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("ginrummy.promptKnockHelp") + "\n")
		case domain.GinRummyPhaseLayoff:
			b.WriteString(i18n.T("ginrummy.promptLayoff") + "\n")
			b.WriteString(i18n.T("ginrummy.promptLayoffHelp") + "\n")
			b.WriteString(i18n.T("ginrummy.promptLayoffSkip") + "\n")
		case domain.GinRummyPhaseRoundEnd:
			b.WriteString(i18n.T("ginrummy.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("ginrummy.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GinRummyCuiPresenter) ActionLogOutput(g interfaces.GinRummyGame) string {
	return actionLogOutputText(g)
}
