package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// presidentPlayerStr returns the display string for a single President player.
func presidentPlayerStr(player *domain.PresidentPlayer, i int) string {
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	if player.GetIsFinished() {
		fmt.Fprintf(&b, ": 上がり (ランク: %s)\n", presidentRankName(player.GetRank()))
	} else {
		fmt.Fprintf(&b, ": %d枚\n", player.GetCardsSize())
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player))
			b.WriteString("\n")
		}
	}
	return b.String()
}

// PresidentCuiPresenter プレジデントCUIプレゼンタークラス
type PresidentCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *PresidentCuiPresenter) Output(pg interfaces.PresidentGame, lastErr error) string {
	return buildCuiOutput("President (プレジデント)", func(b *strings.Builder) {
		for i := 0; i < pg.GetPlayerCnt(); i++ {
			b.WriteString(presidentPlayerStr(pg.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		if pg.GetRevolutionActive() {
			b.WriteString(color.BoldYellow("【革命中】") + "2が最弱、3が最強\n")
		}

		// カード交換記録
		exchangeActions := pg.GetExchangeActions()
		if len(exchangeActions) > 0 {
			b.WriteString(color.Bold("[カード交換]") + "\n")
			for _, ex := range exchangeActions {
				fmt.Fprintf(b, "%s → %s: %s\n",
					cuiPlayerName(pg.GetPlayer(ex.FromPlayerIdx), ex.FromPlayerIdx),
					cuiPlayerName(pg.GetPlayer(ex.ToPlayerIdx), ex.ToPlayerIdx),
					cuiCardSliceStr(ex.Cards))
			}
		}

		// 場のカード
		tableCards := pg.GetTableCards()
		if len(tableCards) > 0 {
			lastPlayIdx := pg.GetLastPlayPlayerIdx()
			fmt.Fprintf(b, "場: %s (出したプレイヤー: %s)\n",
				cuiCardSliceStr(tableCards),
				cuiPlayerName(pg.GetPlayer(lastPlayIdx), lastPlayIdx))
		} else {
			b.WriteString("場: なし (誰でも出せます)\n")
		}

		// 人間の前の行動
		humanAction := pg.GetHumanAction()
		if humanAction != nil {
			if len(humanAction.PlayedCards) == 0 {
				fmt.Fprintf(b, "%sがパスしました\n", cuiPlayerName(pg.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx))
			} else {
				fmt.Fprintf(b, "%sが %s を出しました\n",
					cuiPlayerName(pg.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx),
					cuiCardSliceStr(humanAction.PlayedCards))
			}
		}

		// CPUの行動履歴
		cpuActions := pg.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString(color.Bold("[CPUの行動]") + "\n")
			for _, action := range cpuActions {
				actPlayerName := cuiPlayerName(pg.GetPlayer(action.PlayerIdx), action.PlayerIdx)
				if len(action.PlayedCards) == 0 {
					fmt.Fprintf(b, "%sがパスしました\n", actPlayerName)
				} else {
					fmt.Fprintf(b, "%sが %s を出しました\n", actPlayerName, cuiCardSliceStr(action.PlayedCards))
				}
			}
		}

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		if pg.GetGameEndFlag() {
			b.WriteString("ゲーム終了！\n")
			for i := 0; i < pg.GetPlayerCnt(); i++ {
				player := pg.GetPlayer(i)
				fmt.Fprintf(b, "  %s: %s\n", cuiPlayerName(pg.GetPlayer(i), i), presidentRankName(player.GetRank()))
			}
		} else {
			currentTurn := pg.GetCurrentTurn()
			currentName := cuiPlayerName(pg.GetPlayer(currentTurn), currentTurn)
			fmt.Fprintf(b, "手番: %s\n", currentName)
			b.WriteString("p [インデックス...] でカードを出す / p でパス\n")
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *PresidentCuiPresenter) ActionLogOutput(pg interfaces.PresidentGame) string {
	return actionLogOutputText(pg)
}
