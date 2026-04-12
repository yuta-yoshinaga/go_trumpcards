package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// WarCuiPresenter 戦争CUIプレゼンタークラス
type WarCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *WarCuiPresenter) Output(w interfaces.WarGame, lastErr error) string {
	return buildCuiOutput("War (戦争)", func(b *strings.Builder) {
		cpu := w.GetPlayer(1)
		human := w.GetPlayer(0)

		fmt.Fprintf(b, "CPU: 山札%d枚 / 捨札%d枚 (計%d枚)\n",
			cpu.GetDrawPileSize(), cpu.GetDiscardPileSize(), cpu.TotalCards())

		b.WriteString("----------\n")
		b.WriteString(color.Bold("[場札]") + " ")
		if c := w.GetCpuRevealed(); c != nil {
			fmt.Fprintf(b, "CPU: %s  ", cuiCardStr(c))
		} else {
			b.WriteString("CPU: --  ")
		}
		if c := w.GetPlayerRevealed(); c != nil {
			fmt.Fprintf(b, "あなた: %s", cuiCardStr(c))
		} else {
			b.WriteString("あなた: --")
		}
		fmt.Fprintf(b, "  (場に%d枚)\n", w.GetWarPotSize())
		b.WriteString("----------\n")

		fmt.Fprintf(b, "あなた: 山札%d枚 / 捨札%d枚 (計%d枚)\n",
			human.GetDrawPileSize(), human.GetDiscardPileSize(), human.TotalCards())

		switch w.GetPhase() {
		case domain.WarPhaseReveal:
			b.WriteString("step コマンドで次の1枚をめくります。\n")
		case domain.WarPhaseResolved:
			if w.GetLastWinnerIdx() == 0 {
				b.WriteString(color.Green("あなたがラウンドに勝利！ step で回収します。\n"))
			} else {
				b.WriteString(color.Red("CPUがラウンドに勝利... step で回収します。\n"))
			}
		case domain.WarPhaseWarBury:
			b.WriteString(color.Yellow("戦争発生！ step で伏せ札と表札を出します。\n"))
		}

		if w.GetGameEndFlag() {
			if w.GetWinnerIdx() == 0 {
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
func (p *WarCuiPresenter) ActionLogOutput(w interfaces.WarGame) string {
	return actionLogOutputText(w)
}
