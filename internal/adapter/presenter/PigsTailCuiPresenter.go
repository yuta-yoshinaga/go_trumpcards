package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PigsTailCuiPresenter ぶたのしっぽCUIプレゼンタークラス
type PigsTailCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *PigsTailCuiPresenter) Output(pt interfaces.PigsTailGame, lastErr error) string {
	return buildCuiOutput("Pig's Tail (ぶたのしっぽ)", func(b *strings.Builder) {
		fmt.Fprintf(b, "山札: %d枚 / 場札: %d枚\n", pt.GetCircleCount(), len(pt.GetCenter()))

		topCard := pt.GetCenterTopCard()
		if topCard != nil {
			fmt.Fprintf(b, "場札トップ: %s\n", cuiCardStr(topCard))
		}

		b.WriteString("----------\n")

		for i := 0; i < pt.GetPlayerCnt(); i++ {
			player := pt.GetPlayer(i)
			b.WriteString(cuiPlayerName(player, i))
			fmt.Fprintf(b, ": %d枚\n", player.GetCardsSize())
		}

		b.WriteString("----------\n")

		// CPUの行動履歴を表示
		cpuActions := pt.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString(color.Bold("[CPUの行動]") + "\n")
			for _, action := range cpuActions {
				actPlayerName := cuiPlayerName(pt.GetPlayer(action.DrawPlayerIdx), action.DrawPlayerIdx)
				if action.PenaltyFlag {
					fmt.Fprintf(b, "%sが引いた → ペナルティ！ %d枚引き取り\n", actPlayerName, action.PenaltyCount)
				} else {
					fmt.Fprintf(b, "%sが引いた → セーフ\n", actPlayerName)
				}
			}
		}

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		if pt.GetGameEndFlag() {
			loserIdx := pt.GetLoserIdx()
			if loserIdx >= 0 {
				loserName := cuiPlayerName(pt.GetPlayer(loserIdx), loserIdx)
				loser := pt.GetPlayer(loserIdx)
				fmt.Fprintf(b, "ゲーム終了！ %s (%d枚)\n",
					color.Red(loserName+"の負け！"), loser.GetCardsSize())
			}
		} else {
			currentTurn := pt.GetCurrentTurn()
			currentName := cuiPlayerName(pt.GetPlayer(currentTurn), currentTurn)
			fmt.Fprintf(b, "手番: %s (draw で山札から1枚引く)\n", currentName)
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *PigsTailCuiPresenter) ActionLogOutput(pt interfaces.PigsTailGame) string {
	return actionLogOutputText(pt)
}
