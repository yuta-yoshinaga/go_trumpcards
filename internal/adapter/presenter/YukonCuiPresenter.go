package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// YukonCuiPresenter ユーコンCUIプレゼンタークラス
type YukonCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *YukonCuiPresenter) Output(y interfaces.YukonGame, lastErr error) string {
	return buildCuiOutput("Yukon (ユーコン)", func(b *strings.Builder) {
		// ファンデーション
		b.WriteString("Foundation: ")
		foundation := y.GetFoundation()
		for i := range domain.YukonFoundationCnt {
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
		tableau := y.GetTableau()
		for col := range domain.YukonTableauCnt {
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
		phase := y.GetPhase()
		switch phase {
		case domain.YukonPhasePlaying:
			if y.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", y.GetMoveCount())
		case domain.YukonPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), y.GetMoveCount())
		case domain.YukonPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *YukonCuiPresenter) HintOutput(y interfaces.YukonGame) string {
	hint := y.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: %s", yukonHintStr(hint)) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (p *YukonCuiPresenter) ActionLogOutput(y interfaces.YukonGame) string {
	phase := y.GetPhase()
	if phase == domain.YukonPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(y.GetActionLog())
}

// yukonHintStr ヒントを文字列に変換
func yukonHintStr(hint *domain.YukonHint) string {
	from := fmt.Sprintf("タブロー列%d[%d]", hint.FromCol, hint.CardIndex)
	var to string
	if hint.ToZone == "foundation" {
		to = "ファンデーション"
	} else {
		to = fmt.Sprintf("タブロー列%d", hint.ToCol)
	}
	return fmt.Sprintf("%s → %s", from, to)
}
