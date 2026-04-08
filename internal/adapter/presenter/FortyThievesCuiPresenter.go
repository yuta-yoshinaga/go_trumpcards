package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// fortyThievesColumnStr returns the display string for a FortyThieves tableau column.
func fortyThievesColumnStr(colCards []*domain.FortyThievesTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// FortyThievesCuiPresenter フォーティシーブスCUIプレゼンタークラス
type FortyThievesCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *FortyThievesCuiPresenter) Output(ft interfaces.FortyThievesGame, lastErr error) string {
	return buildCuiOutput("Forty Thieves (フォーティシーブス)", func(b *strings.Builder) {
		// ファンデーション
		b.WriteString("Foundation: ")
		foundation := ft.GetFoundation()
		for i := range domain.FortyThievesFoundationCnt {
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

		// ストックとウェイスト
		fmt.Fprintf(b, "Stock: %d枚", ft.GetStockCount())
		waste := ft.GetWaste()
		if len(waste) > 0 {
			fmt.Fprintf(b, " | Waste: %s", cuiCardStr(waste[len(waste)-1]))
		} else {
			b.WriteString(" | Waste: [空]")
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// タブロー
		tableau := ft.GetTableau()
		for col := range domain.FortyThievesTableauCnt {
			colCards := tableau[col]
			fmt.Fprintf(b, "列%d:", col)
			if len(colCards) == 0 {
				b.WriteString(" [空]")
			} else {
				b.WriteString(fortyThievesColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		phase := ft.GetPhase()
		switch phase {
		case domain.FortyThievesPhasePlaying:
			if ft.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", ft.GetMoveCount())
		case domain.FortyThievesPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), ft.GetMoveCount())
		case domain.FortyThievesPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *FortyThievesCuiPresenter) HintOutput(ft interfaces.FortyThievesGame) string {
	hint := ft.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: %s", fortyThievesHintStr(hint)) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (p *FortyThievesCuiPresenter) ActionLogOutput(ft interfaces.FortyThievesGame) string {
	phase := ft.GetPhase()
	if phase == domain.FortyThievesPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(ft.GetActionLog())
}

// fortyThievesHintStr ヒントを文字列に変換
func fortyThievesHintStr(hint *domain.FortyThievesHint) string {
	var from string
	if hint.FromZone == "tableau" {
		from = fmt.Sprintf("タブロー列%d[%d]", hint.FromCol, hint.CardIndex)
	} else {
		from = "ウェイスト"
	}
	var to string
	if hint.ToZone == "foundation" {
		to = "ファンデーション"
	} else {
		to = fmt.Sprintf("タブロー列%d", hint.ToCol)
	}
	return fmt.Sprintf("%s → %s", from, to)
}
