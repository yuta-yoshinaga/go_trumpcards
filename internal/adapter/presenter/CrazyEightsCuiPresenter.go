package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// crazyEightsPlayerStr returns the display string for a single CrazyEights player.
func crazyEightsPlayerStr(player *domain.CrazyEightsPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("crazyeights.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// CrazyEightsCuiPresenter renders the Crazy Eights CUI view.
type CrazyEightsCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *CrazyEightsCuiPresenter) Output(g interfaces.CrazyEightsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("crazyeights.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("crazyeights.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		// Top of discard pile
		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("crazyeights.discardLine", "card", cuiCardStr(top)))
			if g.GetChosenSuit() > 0 {
				b.WriteString(i18n.Tf("crazyeights.chosenSuit",
					"suit", suitDisplayName(g.GetChosenSuit())))
			}
			b.WriteString("\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(crazyEightsPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("crazyeights.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.CrazyEightsPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("crazyeights.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("crazyeights.promptPlayHelp") + "\n")
			b.WriteString(i18n.T("crazyeights.promptDrawHelp") + "\n")
		case domain.CrazyEightsPhaseChooseSuit:
			b.WriteString(i18n.T("crazyeights.promptChooseSuit") + "\n")
			b.WriteString(i18n.T("crazyeights.promptChooseSuitHelp") + "\n")
		case domain.CrazyEightsPhaseRoundEnd:
			b.WriteString(i18n.T("crazyeights.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("crazyeights.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CrazyEightsCuiPresenter) ActionLogOutput(g interfaces.CrazyEightsGame) string {
	return actionLogOutputText(g)
}

// suitDisplayName returns the suit display string.
func suitDisplayName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	default:
		return "?"
	}
}
