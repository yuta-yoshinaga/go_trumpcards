package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GolfCuiPresenter ゴルフソリティアCUIプレゼンタークラス
type GolfCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (pr *GolfCuiPresenter) Output(g interfaces.GolfGame, lastErr error) string {
	return buildCuiOutput("Golf (ゴルフ)", func(b *strings.Builder) {
		layout := g.GetLayout()

		// タブロー表示 (行ごとに7列分のカードを表示)
		for row := range domain.GolfRowCnt {
			for col := range domain.GolfColCnt {
				if col > 0 {
					b.WriteString("  ")
				}
				gc := layout[col][row]
				if gc == nil || gc.Removed {
					b.WriteString("    ")
				} else if g.IsExposed(col, row) {
					fmt.Fprintf(b, "(%d)%s", col, cuiCardStr(gc.Card))
				} else {
					fmt.Fprintf(b, "   %s", cuiCardStr(gc.Card))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// ストックとウェイスト
		fmt.Fprintf(b, "Stock: %d枚", g.GetStockCount())
		waste := g.GetWaste()
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
		phase := g.GetPhase()
		switch phase {
		case domain.GolfPhasePlaying:
			if g.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", g.GetMoveCount())
		case domain.GolfPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), g.GetMoveCount())
		case domain.GolfPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (pr *GolfCuiPresenter) HintOutput(g interfaces.GolfGame) string {
	hint := g.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: %s\n", golfHintStr(hint))
}

// ActionLogOutput 棋譜をテキスト出力
func (pr *GolfCuiPresenter) ActionLogOutput(g interfaces.GolfGame) string {
	phase := g.GetPhase()
	if phase == domain.GolfPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}

// golfHintStr ヒントを文字列に変換
func golfHintStr(hint *domain.GolfHint) string {
	switch hint.Type {
	case "remove":
		return fmt.Sprintf("カード除去: 列%d", hint.Col)
	case "draw":
		return "ストックからカードを引いてください"
	default:
		return "不明"
	}
}
