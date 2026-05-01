package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// bakersDozenColumnStr returns the display string for a BakersDozen tableau column.
func bakersDozenColumnStr(colCards []*domain.BakersDozenTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// BakersDozenCuiPresenter ベーカーズダズンCUIプレゼンタークラス
type BakersDozenCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *BakersDozenCuiPresenter) Output(bd interfaces.BakersDozenGame, lastErr error) string {
	return buildCuiOutput("Baker's Dozen (ベーカーズ・ダズン)", func(b *strings.Builder) {
		// ファンデーション
		b.WriteString("Foundation: ")
		foundation := bd.GetFoundation()
		for i := range domain.BakersDozenFoundationCnt {
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
		tableau := bd.GetTableau()
		for col := range domain.BakersDozenTableauCnt {
			colCards := tableau[col]
			fmt.Fprintf(b, "列%d:", col)
			if len(colCards) == 0 {
				b.WriteString(" [空]")
			} else {
				b.WriteString(bakersDozenColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		phase := bd.GetPhase()
		switch phase {
		case domain.BakersDozenPhasePlaying:
			if bd.IsStalemate() {
				fmt.Fprintf(b, "%s\n", color.Red("手詰まりです"))
			}
			fmt.Fprintf(b, "手数: %d\n", bd.GetMoveCount())
		case domain.BakersDozenPhaseGameClear:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ゲームクリア！"), bd.GetMoveCount())
		case domain.BakersDozenPhaseGameOver:
			b.WriteString(color.Red("ゲームオーバー") + "\n")
		}
	})
}

// HintOutput ヒントを文字列出力
func (p *BakersDozenCuiPresenter) HintOutput(bd interfaces.BakersDozenGame) string {
	hint := bd.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	return fmt.Sprintf("ヒント: %s", bakersDozenHintStr(hint)) + "\n"
}

// ActionLogOutput 棋譜をテキスト出力
func (p *BakersDozenCuiPresenter) ActionLogOutput(bd interfaces.BakersDozenGame) string {
	phase := bd.GetPhase()
	if phase == domain.BakersDozenPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(bd.GetActionLog())
}

// bakersDozenHintStr ヒントを文字列に変換
func bakersDozenHintStr(hint *domain.BakersDozenHint) string {
	from := fmt.Sprintf("タブロー列%d[%d]", hint.FromCol, hint.CardIndex)
	var to string
	if hint.ToZone == "foundation" {
		to = "ファンデーション"
	} else {
		to = fmt.Sprintf("タブロー列%d", hint.ToCol)
	}
	return fmt.Sprintf("%s → %s", from, to)
}
