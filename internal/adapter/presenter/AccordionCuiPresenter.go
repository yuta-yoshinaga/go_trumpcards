package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AccordionCuiPresenter アコーディオンCUIプレゼンタークラス
type AccordionCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *AccordionCuiPresenter) Output(a interfaces.AccordionGame, lastErr error) string {
	return buildCuiOutput("Accordion (アコーディオン)", func(b *strings.Builder) {
		fmt.Fprintf(b, "残りパイル: %d", a.GetPileCount())
		b.WriteString("\n")
		b.WriteString("----------\n")

		piles := a.GetPiles()
		for i, pile := range piles {
			if len(pile) == 0 {
				continue
			}
			top := pile[len(pile)-1]
			// [idx] 表示カード (山の厚さ)
			if len(pile) == 1 {
				fmt.Fprintf(b, "[%d]%s ", i, cuiCardStr(top))
			} else {
				fmt.Fprintf(b, "[%d]%s(+%d) ", i, cuiCardStr(top), len(pile)-1)
			}
			if (i+1)%8 == 0 {
				b.WriteString("\n")
			}
		}
		b.WriteString("\n----------\n")

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		switch a.GetPhase() {
		case domain.AccordionPhasePlaying:
			if a.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", a.GetMoveCount())
		case domain.AccordionPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), a.GetMoveCount())
		case domain.AccordionPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *AccordionCuiPresenter) HintOutput(a interfaces.AccordionGame) string {
	hint := a.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: パイル%d → パイル%d\n", hint.FromIdx, hint.ToIdx)
}

// ActionLogOutput 棋譜をテキスト出力
func (p *AccordionCuiPresenter) ActionLogOutput(a interfaces.AccordionGame) string {
	if a.GetPhase() == domain.AccordionPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(a.GetActionLog())
}
