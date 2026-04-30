package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// EgyptianRatscrewCuiPresenter エジプシャン・ラットスクリュー CUI プレゼンター
type EgyptianRatscrewCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *EgyptianRatscrewCuiPresenter) Output(g interfaces.EgyptianRatscrewGame, lastErr error) string {
	return buildCuiOutput("Egyptian Ratscrew (エジプシャン・ラットスクリュー)", func(b *strings.Builder) {
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

		if g.GetChanceRemaining() > 0 {
			fmt.Fprintf(b, color.Yellow("チャンスバトル中: 残り %d 回\n"), g.GetChanceRemaining())
		}
		if g.IsSlappable() {
			b.WriteString(color.Yellow("ペア/サンドイッチ成立！ j (slap) で叩いてください\n"))
		} else if g.GetCurrentTurnIdx() == 0 {
			b.WriteString("あなたの番です。s (step) でカードをめくってください\n")
		} else {
			b.WriteString("CPU の番です。tick で進めるか、待ってください\n")
		}

		switch g.GetLastEvent().Kind {
		case domain.EgyptianRatscrewEventSlapCorrect:
			label := "スラップ"
			switch g.GetLastEvent().SlapReason {
			case domain.EgyptianRatscrewSlapReasonPair:
				label = "ペアスラップ"
			case domain.EgyptianRatscrewSlapReasonSandwich:
				label = "サンドイッチスラップ"
			}
			if g.GetLastEvent().PlayerIdx == 0 {
				fmt.Fprintf(b, color.Green("あなたが%sで場札を獲得！\n"), label)
			} else {
				fmt.Fprintf(b, color.Red("CPU が%sで場札を奪った...\n"), label)
			}
		case domain.EgyptianRatscrewEventSlapWrong:
			if g.GetLastEvent().PlayerIdx == 0 {
				b.WriteString(color.Red("誤スラップ！ 1 枚 CPU に渡しました\n"))
			} else {
				b.WriteString(color.Green("CPU が誤スラップ。1 枚もらいました\n"))
			}
		case domain.EgyptianRatscrewEventChanceWin:
			if g.GetLastEvent().PlayerIdx == 0 {
				b.WriteString(color.Green("チャンスバトル勝利！ 場札を獲得\n"))
			} else {
				b.WriteString(color.Red("チャンスバトルで CPU に場札を奪われた...\n"))
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
func (p *EgyptianRatscrewCuiPresenter) ActionLogOutput(g interfaces.EgyptianRatscrewGame) string {
	return actionLogOutputText(g)
}
