package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// tienLenPlayerStr returns the display string for a single Tien Len player.
func tienLenPlayerStr(player *domain.TienLenPlayer, i int) string {
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	if player.GetIsFinished() {
		b.WriteString(i18n.Tf("tienlen.playerFinished", "rank", strconv.Itoa(player.GetRank())) + "\n")
	} else {
		b.WriteString(i18n.Tf("tienlen.playerCardCount", "count", strconv.Itoa(player.GetCardsSize())) + "\n")
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
		}
	}
	return b.String()
}

// TienLenCuiPresenter renders the Tien Len CUI view.
type TienLenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TienLenCuiPresenter) Output(tg interfaces.TienLenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tienlen.helpTitle"), func(b *strings.Builder) {
		for i := 0; i < tg.GetPlayerCnt(); i++ {
			b.WriteString(tienLenPlayerStr(tg.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		if tg.GetTableCards() != nil {
			b.WriteString(i18n.T("tienlen.tableCards") + ": " + cuiCardSliceStr(tg.GetTableCards()) + "\n")
		} else {
			b.WriteString(i18n.T("tienlen.tableEmpty") + "\n")
		}

		if tg.GetGameEndFlag() {
			b.WriteString(i18n.T("tienlen.gameEnd") + "\n")
			for i := 0; i < tg.GetPlayerCnt(); i++ {
				player := tg.GetPlayer(i)
				if player.GetRank() >= 1 {
					var name string
					if player.GetIsHuman() {
						name = i18n.T("tienlen.playerYou")
					} else {
						name = fmt.Sprintf("CPU %d", i)
					}
					fmt.Fprintf(b, "  %s: %s\n", name, i18n.Tf("tienlen.rankN", "rank", strconv.Itoa(player.GetRank())))
				}
			}
		} else {
			for _, action := range tg.GetCpuActions() {
				var actionStr string
				if action.PlayedCards == nil {
					actionStr = i18n.T("tienlen.actionPass")
				} else {
					actionStr = cuiCardSliceStr(action.PlayedCards)
				}
				fmt.Fprintf(b, "CPU %d: %s\n", action.PlayerIdx, actionStr)
			}
			if tg.IsHumanTurn() {
				b.WriteString(i18n.T("tienlen.yourTurn") + "\n")
			}
		}

		if lastErr != nil {
			b.WriteString(lastErr.Error() + "\n")
		}
	})
}

// ActionLogOutput 棋譜を出力
func (p *TienLenCuiPresenter) ActionLogOutput(tg interfaces.TienLenGame) string {
	return actionLogOutputText(tg)
}
