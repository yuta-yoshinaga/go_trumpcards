//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// zhengPlayerStr returns the display string for a single Zheng Shangyou player.
func zhengPlayerStr(player *domain.ZhengPlayer, i int) string {
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	if player.GetIsFinished() {
		b.WriteString(i18n.Tf("zheng.playerFinished", "rank", strconv.Itoa(player.GetRank())) + "\n")
	} else {
		b.WriteString(i18n.Tf("zheng.playerCardCount", "count", strconv.Itoa(player.GetCardsSize())) + "\n")
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
		}
	}
	return b.String()
}

// ZhengCuiPresenter renders the Zheng Shangyou CUI view.
type ZhengCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ZhengCuiPresenter) Output(zg interfaces.ZhengGame, lastErr error) string {
	return buildCuiOutput(i18n.T("zheng.helpTitle"), func(b *strings.Builder) {
		for i := 0; i < zg.GetPlayerCnt(); i++ {
			b.WriteString(zhengPlayerStr(zg.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		if zg.GetTableCards() != nil {
			b.WriteString(i18n.T("zheng.tableCards") + ": " + cuiCardSliceStr(zg.GetTableCards()) + "\n")
		} else {
			b.WriteString(i18n.T("zheng.tableEmpty") + "\n")
		}

		if zg.GetGameEndFlag() {
			b.WriteString(i18n.T("zheng.gameEnd") + "\n")
			for i := 0; i < zg.GetPlayerCnt(); i++ {
				player := zg.GetPlayer(i)
				if player == nil {
					continue
				}
				if player.GetRank() >= 1 {
					// 名前の組み立ては他の表示と同じ cuiPlayerName に任せる。
					// 自前で組むと太字も i18n も抜ける (#4807)。
					name := cuiPlayerName(player, i)
					fmt.Fprintf(b, "  %s: %s\n", name, i18n.Tf("zheng.rankN", "rank", strconv.Itoa(player.GetRank())))
				}
			}
		} else {
			for _, action := range zg.GetCpuActions() {
				var actionStr string
				if action.PlayedCards == nil {
					actionStr = i18n.T("zheng.actionPass")
				} else {
					actionStr = cuiCardSliceStr(action.PlayedCards)
				}
				fmt.Fprintf(b, "%s: %s\n", cuiPlayerName(zg.GetPlayer(action.PlayerIdx), action.PlayerIdx), actionStr)
			}
			if zg.IsHumanTurn() {
				b.WriteString(i18n.T("zheng.yourTurn") + "\n")
				// Combo-strength rules, so CLI players can judge a legal play
				// (and self-diagnose an invalid combo shown in the error line).
				b.WriteString(i18n.T("zheng.comboRulesHint") + "\n")
			}
		}

		if lastErr != nil {
			b.WriteString(lastErr.Error() + "\n")
		}
	})
}

// ActionLogOutput 棋譜を出力
func (p *ZhengCuiPresenter) ActionLogOutput(zg interfaces.ZhengGame) string {
	return actionLogOutputText(zg)
}
