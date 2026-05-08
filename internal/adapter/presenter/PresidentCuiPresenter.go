package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// presidentPlayerStr returns the display string for a single President player.
func presidentPlayerStr(player *domain.PresidentPlayer, i int) string {
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	if player.GetIsFinished() {
		b.WriteString(i18n.Tf("president.playerFinished",
			"rank", presidentRankName(player.GetRank())) + "\n")
	} else {
		b.WriteString(i18n.Tf("president.playerHand",
			"count", strconv.Itoa(player.GetCardsSize())) + "\n")
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
		}
	}
	return b.String()
}

// PresidentCuiPresenter renders the President / Scum CUI view.
type PresidentCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *PresidentCuiPresenter) Output(pg interfaces.PresidentGame, lastErr error) string {
	return buildCuiOutput(i18n.T("president.helpTitle"), func(b *strings.Builder) {
		for i := 0; i < pg.GetPlayerCnt(); i++ {
			b.WriteString(presidentPlayerStr(pg.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		if pg.GetRevolutionActive() {
			b.WriteString(color.BoldYellow(i18n.T("president.revolutionLabel")) +
				i18n.T("president.revolutionSuffix") + "\n")
		}

		// Card exchange records
		if exchangeActions := pg.GetExchangeActions(); len(exchangeActions) > 0 {
			b.WriteString(color.Bold(i18n.T("president.exchangesHeader")) + "\n")
			for _, ex := range exchangeActions {
				b.WriteString(i18n.Tf("president.exchangeLine",
					"from", cuiPlayerName(pg.GetPlayer(ex.FromPlayerIdx), ex.FromPlayerIdx),
					"to", cuiPlayerName(pg.GetPlayer(ex.ToPlayerIdx), ex.ToPlayerIdx),
					"cards", cuiCardSliceStr(ex.Cards)) + "\n")
			}
		}

		// Table cards
		if tableCards := pg.GetTableCards(); len(tableCards) > 0 {
			lastPlayIdx := pg.GetLastPlayPlayerIdx()
			b.WriteString(i18n.Tf("president.tableLine",
				"cards", cuiCardSliceStr(tableCards),
				"name", cuiPlayerName(pg.GetPlayer(lastPlayIdx), lastPlayIdx)) + "\n")
		} else {
			b.WriteString(i18n.T("president.tableEmpty") + "\n")
		}

		// Last human action
		if humanAction := pg.GetHumanAction(); humanAction != nil {
			actName := cuiPlayerName(pg.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx)
			if len(humanAction.PlayedCards) == 0 {
				b.WriteString(i18n.Tf("president.actionPass", "name", actName) + "\n")
			} else {
				b.WriteString(i18n.Tf("president.actionPlay",
					"name", actName,
					"cards", cuiCardSliceStr(humanAction.PlayedCards)) + "\n")
			}
		}

		// CPU action history
		if cpuActions := pg.GetCpuActions(); len(cpuActions) > 0 {
			b.WriteString(color.Bold(i18n.T("president.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				actName := cuiPlayerName(pg.GetPlayer(action.PlayerIdx), action.PlayerIdx)
				if len(action.PlayedCards) == 0 {
					b.WriteString(i18n.Tf("president.actionPass", "name", actName) + "\n")
				} else {
					b.WriteString(i18n.Tf("president.actionPlay",
						"name", actName,
						"cards", cuiCardSliceStr(action.PlayedCards)) + "\n")
				}
			}
		}

		cuiErrorBlock(b, lastErr)

		if pg.GetGameEndFlag() {
			b.WriteString(i18n.T("president.gameEnd") + "\n")
			for i := 0; i < pg.GetPlayerCnt(); i++ {
				player := pg.GetPlayer(i)
				b.WriteString(i18n.Tf("president.rankEntry",
					"name", cuiPlayerName(pg.GetPlayer(i), i),
					"rank", presidentRankName(player.GetRank())) + "\n")
			}
			return
		}
		currentTurn := pg.GetCurrentTurn()
		b.WriteString(i18n.Tf("president.promptCurrentTurn",
			"name", cuiPlayerName(pg.GetPlayer(currentTurn), currentTurn)) + "\n")
		b.WriteString(i18n.T("president.promptPlayHelp") + "\n")
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PresidentCuiPresenter) ActionLogOutput(pg interfaces.PresidentGame) string {
	return actionLogOutputText(pg)
}
