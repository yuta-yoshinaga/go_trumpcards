package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// thirtyOnePlayerStr returns the display string for a single ThirtyOne player.
func thirtyOnePlayerStr(g interfaces.ThirtyOneGame, player *domain.ThirtyOnePlayer, i int) string {
	var b strings.Builder
	reveal := thirtyOneReveal(g)
	score := "?"
	if player.GetIsHuman() || reveal {
		score = strconv.Itoa(player.BestSuitScore())
	}
	if player.IsEliminated() {
		score = "OUT"
	}
	b.WriteString(i18n.Tf("thirtyone.playerLine",
		"name", cuiPlayerName(player, i),
		"lives", strconv.Itoa(player.GetLives()),
		"score", score,
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// ThirtyOneCuiPresenter renders the ThirtyOne CUI view.
type ThirtyOneCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ThirtyOneCuiPresenter) Output(g interfaces.ThirtyOneGame, lastErr error) string {
	return buildCuiOutput(i18n.T("thirtyone.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("thirtyone.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("thirtyone.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(thirtyOnePlayerStr(g, g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("thirtyone.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(g.GetWinnerIdx()), g.GetWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.ThirtyOnePhaseDraw:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("thirtyone.promptDraw", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("thirtyone.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("thirtyone.promptDrawHelpDiscard") + "\n")
			if g.GetKnockerIdx() < 0 {
				b.WriteString(i18n.T("thirtyone.promptKnockHelp") + "\n")
			}
		case domain.ThirtyOnePhaseDiscard:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("thirtyone.promptDiscard", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("thirtyone.promptDiscardHelp") + "\n")
		case domain.ThirtyOnePhaseRoundEnd:
			if g.GetThirtyOneIdx() >= 0 {
				b.WriteString(i18n.Tf("thirtyone.promptThirtyOne",
					"name", cuiPlayerName(g.GetPlayer(g.GetThirtyOneIdx()), g.GetThirtyOneIdx())) + "\n")
			}
			b.WriteString(i18n.T("thirtyone.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("thirtyone.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ThirtyOneCuiPresenter) ActionLogOutput(g interfaces.ThirtyOneGame) string {
	return actionLogOutputText(g)
}
