package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// goFishPlayerStr returns the display string for a single GoFish player.
func goFishPlayerStr(player *domain.GoFishPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("gofish.playerLine",
		"name", cuiPlayerName(player, i),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"books", strconv.Itoa(player.GetBookCount())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// GoFishCuiPresenter renders the Go Fish CUI view.
type GoFishCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *GoFishCuiPresenter) Output(gf interfaces.GoFishGame, lastErr error) string {
	return buildCuiOutput(i18n.T("gofish.outputTitle"), func(b *strings.Builder) {
		for i := 0; i < gf.GetPlayerCnt(); i++ {
			b.WriteString(goFishPlayerStr(gf.GetPlayer(i), i))
		}

		b.WriteString("----------\n")
		b.WriteString(i18n.Tf("gofish.deckLine",
			"count", strconv.Itoa(gf.GetDeckRemaining())) + "\n")

		// Last ask result
		if gf.GetLastAskPlayerIdx() >= 0 {
			askerName := cuiPlayerName(gf.GetPlayer(gf.GetLastAskPlayerIdx()), gf.GetLastAskPlayerIdx())
			targetName := cuiPlayerName(gf.GetPlayer(gf.GetLastAskTargetIdx()), gf.GetLastAskTargetIdx())
			rankStr := strconv.Itoa(gf.GetLastAskRank())
			if gf.GetLastAskSuccess() {
				b.WriteString(i18n.Tf("gofish.askSuccess",
					"asker", askerName,
					"target", targetName,
					"rank", rankStr,
					"count", strconv.Itoa(len(gf.GetLastCardsReceived()))) + "\n")
			} else {
				b.WriteString(i18n.Tf("gofish.askFail",
					"asker", askerName,
					"target", targetName,
					"rank", rankStr) + "\n")
			}
			if gf.GetLastBookFormed() {
				b.WriteString(i18n.Tf("gofish.bookFormed",
					"rank", strconv.Itoa(gf.GetLastBookRank())) + "\n")
			}
		}

		// CPU action history
		for _, action := range gf.GetCpuActions() {
			askerName := cuiPlayerName(gf.GetPlayer(action.AskPlayerIdx), action.AskPlayerIdx)
			targetName := cuiPlayerName(gf.GetPlayer(action.AskTargetIdx), action.AskTargetIdx)
			rankStr := strconv.Itoa(action.AskRank)
			if action.Success {
				b.WriteString(i18n.Tf("gofish.cpuAskSuccess",
					"asker", askerName,
					"target", targetName,
					"rank", rankStr,
					"count", strconv.Itoa(action.CardsReceived)) + "\n")
			} else {
				b.WriteString(i18n.Tf("gofish.cpuAskFail",
					"asker", askerName,
					"target", targetName,
					"rank", rankStr) + "\n")
			}
			if action.BookFormed {
				b.WriteString(i18n.Tf("gofish.cpuBookFormed",
					"asker", askerName,
					"rank", strconv.Itoa(action.BookRank)) + "\n")
			}
		}

		b.WriteString("----------\n")

		if lastErr != nil {
			b.WriteString(i18n.Tf("gofish.errorLine", "msg", lastErr.Error()) + "\n")
		}

		if gf.GetGameEndFlag() {
			winnerIdx := gf.GetWinnerIdx()
			winner := gf.GetPlayer(winnerIdx)
			if winner != nil && winner.GetIsHuman() {
				b.WriteString(i18n.T("gofish.winHuman") + "\n")
			} else {
				b.WriteString(i18n.Tf("gofish.winCpu",
					"idx", strconv.Itoa(winnerIdx)) + "\n")
			}
			for i := 0; i < gf.GetPlayerCnt(); i++ {
				b.WriteString(i18n.Tf("gofish.scoreEntry",
					"name", cuiPlayerName(gf.GetPlayer(i), i),
					"count", strconv.Itoa(gf.GetPlayer(i).GetBookCount())) + "\n")
			}
			return
		}
		turnName := cuiPlayerName(gf.GetPlayer(gf.GetCurrentTurn()), gf.GetCurrentTurn())
		b.WriteString(i18n.Tf("gofish.promptCurrentTurn", "name", turnName) + "\n")
		if gf.IsHumanTurn() {
			b.WriteString(i18n.T("gofish.promptHumanHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GoFishCuiPresenter) ActionLogOutput(gf interfaces.GoFishGame) string {
	return actionLogOutputText(gf)
}
