package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TriPeaksCuiPresenter トリピークスCUIプレゼンタークラス
type TriPeaksCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (pr *TriPeaksCuiPresenter) Output(t interfaces.TriPeaksGame, lastErr error) string {
	return buildCuiOutput("TriPeaks (トリピークス)", func(b *strings.Builder) {
		layout := t.GetLayout()

		// タブロー表示
		for row := range domain.TriPeaksRowCnt {
			// インデント
			indent := strings.Repeat("  ", domain.TriPeaksRowCnt-1-row)
			b.WriteString(indent)
			first := true
			for col := range domain.TriPeaksColCnt {
				tc := layout[row][col]
				if tc == nil {
					continue
				}
				if !first {
					b.WriteString("  ")
				}
				first = false
				if tc.Removed {
					b.WriteString("    ")
				} else {
					fmt.Fprintf(b, "(%d,%d)%s", row, col, cuiCardStr(tc.Card))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// ストックとウェイスト
		fmt.Fprintf(b, "Stock: %d枚", t.GetStockCount())
		waste := t.GetWaste()
		if len(waste) > 0 {
			fmt.Fprintf(b, " | Waste: %s", cuiCardStr(waste[len(waste)-1]))
		} else {
			b.WriteString(" | Waste: [空]")
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		phase := t.GetPhase()
		switch phase {
		case domain.TriPeaksPhasePlaying:
			if t.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", t.GetMoveCount())
		case domain.TriPeaksPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), t.GetMoveCount())
		case domain.TriPeaksPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (pr *TriPeaksCuiPresenter) HintOutput(t interfaces.TriPeaksGame) string {
	hint := t.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: %s\n", triPeaksHintStr(hint))
}

// ActionLogOutput 棋譜をテキスト出力
func (pr *TriPeaksCuiPresenter) ActionLogOutput(t interfaces.TriPeaksGame) string {
	phase := t.GetPhase()
	if phase == domain.TriPeaksPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(t.GetActionLog())
}

// triPeaksHintStr ヒントを文字列に変換
func triPeaksHintStr(hint *domain.TriPeaksHint) string {
	switch hint.Type {
	case "remove":
		return fmt.Sprintf("カード除去: (%d,%d)", hint.Row, hint.Col)
	case "draw":
		return "ストックからカードを引いてください"
	default:
		return "不明"
	}
}
