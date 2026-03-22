package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FreeCellCuiPresenter フリーセルCUIプレゼンタークラス
type FreeCellCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *FreeCellCuiPresenter) Output(f interfaces.FreeCellGame, lastErr error) string {
	return buildCuiOutput("FreeCell (フリーセル)", func(b *strings.Builder) {
		// フリーセル
		b.WriteString("FreeCells: ")
		freeCells := f.GetFreeCells()
		for i := 0; i < domain.FreeCellCellCnt; i++ {
			if i != 0 {
				b.WriteString(" | ")
			}
			if freeCells[i] == nil {
				b.WriteString("[空]")
			} else {
				b.WriteString(cuiCardStr(freeCells[i]))
			}
		}
		b.WriteString("\n")

		// ファンデーション
		b.WriteString("Foundation: ")
		foundation := f.GetFoundation()
		for i := 0; i < domain.FreeCellFoundationCnt; i++ {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString("[空]")
			} else {
				topCard := pile[len(pile)-1]
				b.WriteString(cuiCardStr(topCard))
			}
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// タブロー
		tableau := f.GetTableau()
		for col := 0; col < domain.FreeCellTableauCnt; col++ {
			colCards := tableau[col]
			fmt.Fprintf(b, "列%d:", col)
			if len(colCards) == 0 {
				b.WriteString(" [空]")
			} else {
				for j, c := range colCards {
					fmt.Fprintf(b, " [%d]%s", j, cuiCardStr(c))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		phase := f.GetPhase()
		switch phase {
		case domain.FreeCellPhasePlaying:
			if f.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", f.GetMoveCount())
		case domain.FreeCellPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), f.GetMoveCount())
		case domain.FreeCellPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *FreeCellCuiPresenter) HintOutput(f interfaces.FreeCellGame) string {
	hint := f.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: %s", freeCellHintStr(hint)) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (p *FreeCellCuiPresenter) ActionLogOutput(f interfaces.FreeCellGame) string {
	phase := f.GetPhase()
	if phase == domain.FreeCellPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(f.GetActionLog())
}

// freeCellHintStr ヒントを文字列に変換
func freeCellHintStr(hint *domain.FreeCellHint) string {
	var from string
	switch hint.FromZone {
	case "tableau":
		from = fmt.Sprintf("タブロー列%d[%d]", hint.FromCol, hint.CardIndex)
	case "freecell":
		from = fmt.Sprintf("フリーセル%d", hint.FromCol)
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = "ファンデーション"
	case "tableau":
		to = fmt.Sprintf("タブロー列%d", hint.ToCol)
	case "freecell":
		to = fmt.Sprintf("フリーセル%d", hint.ToCol)
	}
	return fmt.Sprintf("%s → %s", from, to)
}
