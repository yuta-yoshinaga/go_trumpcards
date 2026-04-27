package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SlapjackCuiPresenter スラップジャック CUI プレゼンター
type SlapjackCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *SlapjackCuiPresenter) Output(g interfaces.SlapjackGame, lastErr error) string {
	return buildCuiOutput("Slapjack (スラップジャック)", func(b *strings.Builder) {
		cpu := g.GetPlayer(1)
		human := g.GetPlayer(0)

		fmt.Fprintf(b, "CPU: ストック%d枚\n", cpu.GetStockSize())
		b.WriteString("----------\n")
		b.WriteString(color.Bold("[場札]") + " ")
		if c := g.GetTopCard(); c != nil {
			fmt.Fprintf(b, "%s  (場に%d枚)\n", cuiCardStr(c), g.GetCenterPileSize())
		} else {
			fmt.Fprintf(b, "--  (場に%d枚)\n", g.GetCenterPileSize())
		}
		b.WriteString("----------\n")
		fmt.Fprintf(b, "あなた: ストック%d枚\n", human.GetStockSize())

		if g.IsTopJack() {
			b.WriteString(color.Yellow("J が場に出ました！ j (slap) で叩いてください\n"))
		} else if g.GetCurrentTurnIdx() == 0 {
			b.WriteString("あなたの番です。s (step) でカードをめくってください\n")
		} else {
			b.WriteString("CPU の番です。tick で進めるか、待ってください\n")
		}

		switch g.GetLastEvent().Kind {
		case domain.SlapjackEventSlapCorrect:
			if g.GetLastEvent().PlayerIdx == 0 {
				b.WriteString(color.Green("あなたが正しくスラップ! 場札を獲得\n"))
			} else {
				b.WriteString(color.Red("CPU が先にスラップ... 場札を奪われた\n"))
			}
		case domain.SlapjackEventSlapWrong:
			if g.GetLastEvent().PlayerIdx == 0 {
				b.WriteString(color.Red("誤スラップ！ 1 枚 CPU に渡しました\n"))
			} else {
				b.WriteString(color.Green("CPU が誤スラップ。1 枚もらいました\n"))
			}
		}

		if g.GetGameEndFlag() {
			if g.GetWinnerIdx() == 0 {
				b.WriteString(color.Green("あなたの勝ちです！\n"))
			} else {
				b.WriteString(color.Red("CPUの勝ちです...\n"))
			}
		}

		if lastErr != nil {
			fmt.Fprintf(b, "%s %s\n", color.Red("[エラー]"), lastErr.Error())
		}
	})
}

// ActionLogOutput 棋譜を文字列出力
func (p *SlapjackCuiPresenter) ActionLogOutput(g interfaces.SlapjackGame) string {
	return actionLogOutputText(g)
}
