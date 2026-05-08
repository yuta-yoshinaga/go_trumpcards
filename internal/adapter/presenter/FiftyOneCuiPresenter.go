package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// FiftyOneCuiPresenter renders the Fifty-one CUI view.
type FiftyOneCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *FiftyOneCuiPresenter) Output(fo interfaces.FiftyOneGame, lastErr error) string {
	return buildCuiOutput(i18n.T("fiftyone.outputTitle"), func(b *strings.Builder) {
		// Per-player info
		for i := 0; i < fo.GetPlayerCnt(); i++ {
			player := fo.GetPlayer(i)
			b.WriteString(cuiPlayerName(player, i))
			score := player.BestSuitScore()
			if player.GetIsHuman() {
				b.WriteString(i18n.Tf("fiftyone.humanScoreLine",
					"score", strconv.Itoa(score)) + "\n")
				b.WriteString(cuiIndexedCardListStr(player) + "\n")
			} else {
				scoreStr := i18n.T("fiftyone.scoreUnknown")
				if fo.GetGameEndFlag() {
					scoreStr = strconv.Itoa(score)
				}
				b.WriteString(i18n.Tf("fiftyone.cpuScoreLine",
					"count", strconv.Itoa(player.GetCardsSize()),
					"score", scoreStr) + "\n")
			}
		}

		// Table cards
		b.WriteString("----------\n")
		b.WriteString(i18n.T("fiftyone.tableHeader"))
		tableCards := fo.GetTableCards()
		for i, c := range tableCards {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString("[" + strconv.Itoa(i) + "]" + cuiCardStr(c))
		}
		b.WriteString("\n")

		// Stop state
		if fo.GetStopCallerIdx() >= 0 {
			callerName := cuiPlayerName(fo.GetPlayer(fo.GetStopCallerIdx()), fo.GetStopCallerIdx())
			b.WriteString(i18n.Tf("fiftyone.stopCalled", "name", callerName) + "\n")
		}

		b.WriteString("----------\n")

		if lastErr != nil {
			b.WriteString(i18n.Tf("fiftyone.errorLine", "msg", lastErr.Error()) + "\n")
		}

		if fo.GetGameEndFlag() {
			b.WriteString(i18n.T("fiftyone.gameEndHeader") + "\n")
			for i := 0; i < fo.GetPlayerCnt(); i++ {
				name := cuiPlayerName(fo.GetPlayer(i), i)
				b.WriteString(i18n.Tf("fiftyone.scoreEntry",
					"name", name,
					"score", strconv.Itoa(fo.GetPlayer(i).BestSuitScore())) + "\n")
			}
			winnerIdx := fo.GetWinnerIdx()
			winner := fo.GetPlayer(winnerIdx)
			if winner != nil && winner.GetIsHuman() {
				b.WriteString(i18n.T("fiftyone.winHuman") + "\n")
			} else {
				b.WriteString(i18n.Tf("fiftyone.winCpu",
					"idx", strconv.Itoa(winnerIdx)) + "\n")
			}
		} else if fo.IsHumanTurn() {
			b.WriteString(i18n.T("fiftyone.promptHumanTurn") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FiftyOneCuiPresenter) ActionLogOutput(fo interfaces.FiftyOneGame) string {
	return actionLogOutputText(fo)
}
