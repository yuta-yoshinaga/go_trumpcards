package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CalculationCuiPresenter カルキュレーションCUIプレゼンタークラス
type CalculationCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *CalculationCuiPresenter) Output(g interfaces.CalculationGame, lastErr error) string {
	return buildCuiOutput("Calculation (カルキュレーション)", func(b *strings.Builder) {
		// ファンデーション
		foundations := g.GetFoundations()
		stepLabels := []string{"+1", "+2", "+3", "+4"}
		for i := range domain.CalculationFoundationCnt {
			pile := foundations[i]
			fmt.Fprintf(b, "[F%d %s] ", i, stepLabels[i])
			if len(pile) == 0 {
				b.WriteString("(empty)")
			} else {
				top := pile[len(pile)-1]
				fmt.Fprintf(b, "%s (%d/%d)", cuiCardStr(top), len(pile), domain.CardValueMax)
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// ウェイスト
		wastes := g.GetWastes()
		for i := range domain.CalculationWasteCnt {
			pile := wastes[i]
			fmt.Fprintf(b, "[W%d] ", i)
			if len(pile) == 0 {
				b.WriteString("(empty)")
			} else {
				top := pile[len(pile)-1]
				fmt.Fprintf(b, "%s (%d枚)", cuiCardStr(top), len(pile))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// ストック
		fmt.Fprintf(b, "ストック: %d枚", g.GetStockCount())
		if top := g.GetStockTop(); top != nil {
			fmt.Fprintf(b, " 次のカード: %s", cuiCardStr(top))
		}
		b.WriteString("\n")

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		switch g.GetPhase() {
		case domain.CalculationPhasePlaying:
			if g.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", g.GetMoveCount())
		case domain.CalculationPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), g.GetMoveCount())
		case domain.CalculationPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *CalculationCuiPresenter) HintOutput(g interfaces.CalculationGame) string {
	hint := g.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	switch hint.FromZone {
	case "stock":
		return fmt.Sprintf("ヒント: ストック → ファンデーション%d\n", hint.FoundationIdx)
	case "waste":
		return fmt.Sprintf("ヒント: ウェイスト%d → ファンデーション%d\n", hint.WasteIdx, hint.FoundationIdx)
	}
	return "ヒントはありません。\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (p *CalculationCuiPresenter) ActionLogOutput(g interfaces.CalculationGame) string {
	if g.GetPhase() == domain.CalculationPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
