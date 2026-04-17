package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ScorpionCuiPresenter スコーピオンCUIプレゼンタークラス
type ScorpionCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *ScorpionCuiPresenter) Output(s interfaces.ScorpionGame, lastErr error) string {
	return buildCuiOutput("Scorpion (スコーピオン)", func(b *strings.Builder) {
		fmt.Fprintf(b, "Completed: %d/%d", s.GetCompletedSuits(), domain.ScorpionCompletedSuitsCnt)
		fmt.Fprintf(b, " | Stock: %d枚", s.GetStockCount())
		b.WriteString("\n")

		b.WriteString("----------\n")

		tableau := s.GetTableau()
		for col := range domain.ScorpionTableauCnt {
			colCards := tableau[col]
			fmt.Fprintf(b, "列%d:", col)
			if len(colCards) == 0 {
				b.WriteString(" [空]")
			} else {
				b.WriteString(klondikeColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		phase := s.GetPhase()
		switch phase {
		case domain.ScorpionPhasePlaying:
			if s.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", s.GetMoveCount())
		case domain.ScorpionPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), s.GetMoveCount())
		case domain.ScorpionPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *ScorpionCuiPresenter) HintOutput(s interfaces.ScorpionGame) string {
	hint := s.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	if hint.FromCol < 0 {
		return "ヒント: d でストックから配ってください\n"
	}
	return fmt.Sprintf("ヒント: タブロー列%d[%d] → タブロー列%d\n", hint.FromCol, hint.CardIndex, hint.ToCol)
}

// ActionLogOutput 棋譜をテキスト出力
func (p *ScorpionCuiPresenter) ActionLogOutput(s interfaces.ScorpionGame) string {
	phase := s.GetPhase()
	if phase == domain.ScorpionPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
