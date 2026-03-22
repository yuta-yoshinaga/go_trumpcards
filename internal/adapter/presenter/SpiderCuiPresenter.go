package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// spiderColumnStr returns the display string for a Spider tableau column.
func spiderColumnStr(colCards []*domain.SpiderTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		if tc.FaceUp {
			parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
		} else {
			parts[j] = fmt.Sprintf(" [%d]??", j)
		}
	}
	return strings.Join(parts, " ")
}

// SpiderCuiPresenter スパイダーソリティアCUIプレゼンタークラス
type SpiderCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *SpiderCuiPresenter) Output(s interfaces.SpiderGame, lastErr error) string {
	return buildCuiOutput("Spider Solitaire (スパイダーソリティア)", func(b *strings.Builder) {
		// 完成スート数
		fmt.Fprintf(b, "Completed: %d/%d", s.GetCompletedSuits(), domain.SpiderFoundationCnt)
		fmt.Fprintf(b, " | Stock: %d枚", s.GetStockCount())
		fmt.Fprintf(b, " | Score: %d", s.GetScore())
		b.WriteString("\n")

		b.WriteString("----------\n")

		// タブロー
		tableau := s.GetTableau()
		for col := range domain.SpiderTableauCnt {
			colCards := tableau[col]
			fmt.Fprintf(b, "列%d:", col)
			if len(colCards) == 0 {
				b.WriteString(" [空]")
			} else {
				b.WriteString(spiderColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		phase := s.GetPhase()
		switch phase {
		case domain.SpiderPhasePlaying:
			if s.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", s.GetMoveCount())
		case domain.SpiderPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d スコア: %d\n", color.Green("ゲームクリア！"), s.GetMoveCount(), s.GetScore())
		case domain.SpiderPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *SpiderCuiPresenter) HintOutput(s interfaces.SpiderGame) string {
	hint := s.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: タブロー列%d[%d] → タブロー列%d\n", hint.FromCol, hint.CardIndex, hint.ToCol)
}

// ActionLogOutput 棋譜をテキスト出力
func (p *SpiderCuiPresenter) ActionLogOutput(s interfaces.SpiderGame) string {
	phase := s.GetPhase()
	if phase == domain.SpiderPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
