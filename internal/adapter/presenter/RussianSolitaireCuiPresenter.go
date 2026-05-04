package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RussianSolitaireCuiPresenter ロシアンソリティアCUIプレゼンタークラス
type RussianSolitaireCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *RussianSolitaireCuiPresenter) Output(r interfaces.RussianSolitaireGame, lastErr error) string {
	return buildCuiOutput("Russian Solitaire (ロシアンソリティア)", func(b *strings.Builder) {
		// ファンデーション
		b.WriteString("Foundation: ")
		foundation := r.GetFoundation()
		for i := range domain.RussianSolitaireFoundationCnt {
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
		tableau := r.GetTableau()
		for col := range domain.RussianSolitaireTableauCnt {
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

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		phase := r.GetPhase()
		switch phase {
		case domain.RussianSolitairePhasePlaying:
			if r.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", r.GetMoveCount())
		case domain.RussianSolitairePhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), r.GetMoveCount())
		case domain.RussianSolitairePhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *RussianSolitaireCuiPresenter) HintOutput(r interfaces.RussianSolitaireGame) string {
	hint := r.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: %s", russianSolitaireHintStr(hint)) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (p *RussianSolitaireCuiPresenter) ActionLogOutput(r interfaces.RussianSolitaireGame) string {
	phase := r.GetPhase()
	if phase == domain.RussianSolitairePhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(r.GetActionLog())
}

// russianSolitaireHintStr ヒントを文字列に変換
func russianSolitaireHintStr(hint *domain.RussianSolitaireHint) string {
	from := fmt.Sprintf("タブロー列%d[%d]", hint.FromCol, hint.CardIndex)
	var to string
	if hint.ToZone == "foundation" {
		to = "ファンデーション"
	} else {
		to = fmt.Sprintf("タブロー列%d", hint.ToCol)
	}
	return fmt.Sprintf("%s → %s", from, to)
}
