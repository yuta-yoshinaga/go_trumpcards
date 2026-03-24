package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PyramidCuiPresenter ピラミッドCUIプレゼンタークラス
type PyramidCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (pr *PyramidCuiPresenter) Output(p interfaces.PyramidGame, lastErr error) string {
	return buildCuiOutput("Pyramid (ピラミッド)", func(b *strings.Builder) {
		// ピラミッド表示
		pyramid := p.GetPyramid()
		for row := range domain.PyramidRowCnt {
			// インデント（三角形の形にする）
			indent := strings.Repeat("  ", domain.PyramidRowCnt-1-row)
			b.WriteString(indent)
			for col := range row + 1 {
				if col > 0 {
					b.WriteString("  ")
				}
				pc := pyramid[row][col]
				if pc.Removed {
					b.WriteString("    ")
				} else {
					b.WriteString(fmt.Sprintf("(%d,%d)%s", row, col, cuiCardStr(pc.Card)))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// ストックとウェイスト
		fmt.Fprintf(b, "Stock: %d枚", p.GetStockCount())
		waste := p.GetWaste()
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
		phase := p.GetPhase()
		switch phase {
		case domain.PyramidPhasePlaying:
			if p.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", p.GetMoveCount())
		case domain.PyramidPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), p.GetMoveCount())
		case domain.PyramidPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (pr *PyramidCuiPresenter) HintOutput(p interfaces.PyramidGame) string {
	hint := p.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: %s\n", pyramidHintStr(hint))
}

// ActionLogOutput 棋譜をテキスト出力
func (pr *PyramidCuiPresenter) ActionLogOutput(p interfaces.PyramidGame) string {
	phase := p.GetPhase()
	if phase == domain.PyramidPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(p.GetActionLog())
}

// pyramidHintStr ヒントを文字列に変換
func pyramidHintStr(hint *domain.PyramidHint) string {
	switch hint.Type {
	case "king":
		return fmt.Sprintf("キング除去: (%d,%d)", hint.Row1, hint.Col1)
	case "pair":
		return fmt.Sprintf("ペア除去: (%d,%d)+(%d,%d)", hint.Row1, hint.Col1, hint.Row2, hint.Col2)
	case "waste_king":
		return "ウェイストのキング除去"
	case "waste_pair":
		return fmt.Sprintf("ウェイスト+ピラミッド(%d,%d)", hint.Row1, hint.Col1)
	default:
		return "不明"
	}
}
