//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// bigbenColumnStr returns the display string for one tableau column.
func bigBenColumnStr(colCards []*domain.BigBenTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// BigBenCuiPresenter renders the Big Ben CUI view.
type BigBenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BigBenCuiPresenter) Output(gc interfaces.BigBenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bigben.helpTitle"), func(b *strings.Builder) {
		// 文字盤。時刻・現在の最上段・目標ランク・完成印を 1 行ずつ出す。
		// 12 個を 1 行に詰めると目標ランクが読めない。
		b.WriteString(i18n.T("bigben.foundationHeader") + "\n")
		foundation := gc.GetFoundation()
		for i := range domain.BigBenFoundationCnt {
			b.WriteString(i18n.Tf("bigben.faceLabel", "idx", strconv.Itoa(i)))
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(" " + cuiCardStr(pile[len(pile)-1]))
			}
			// ランクは数字のまま。cuiCardStr が "SPADE 5" と出す以上、ここだけ
			// A/J/Q/K の表記にすると同じ画面で二通りの読み方が混ざる。
			b.WriteString(" " + i18n.Tf("bigben.faceTarget",
				"rank", strconv.Itoa(domain.BigBenTargetRank(i))))
			if gc.IsFoundationComplete(i) {
				b.WriteString(" " + color.Green(i18n.T("bigben.faceComplete")))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// **山札の残りを出す。**補充がこのゲームの逃げ道なので、あと何枚あるかが
		// 読めないと「詰んだのか、配れば動くのか」が分からない。
		b.WriteString(i18n.Tf("bigben.stockLine", "count", strconv.Itoa(gc.GetStockCount())) + "\n")

		// タブロー
		tableau := gc.GetTableau()
		for col := range domain.BigBenTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("bigben.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(bigBenColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch gc.GetPhase() {
		case domain.BigBenPhasePlaying:
			if gc.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, and the command
				// to use, matching the web StalemateEscapeButton.
				if n := gc.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("bigben.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(gc.GetMoveCount())) + "\n")
		case domain.BigBenPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(gc.GetMoveCount())) + "\n")
		case domain.BigBenPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			faces := 0
			for i := range gc.GetFoundation() {
				if gc.IsFoundationComplete(i) {
					faces++
				}
			}
			b.WriteString(color.Yellow(cuiSolitaireGameOverFaces(
				faces, domain.BigBenFoundationCnt)) + "\n")
		}
	})
}

// HintOutput emits the current Big Ben hint.
func (p *BigBenCuiPresenter) HintOutput(gc interfaces.BigBenGame) string {
	hint := gc.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("bigben.hintFrom", "col", strconv.Itoa(hint.FromCol))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.Tf("bigben.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	} else {
		to = i18n.Tf("bigben.hintToTableau", "col", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("bigben.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BigBenCuiPresenter) ActionLogOutput(gc interfaces.BigBenGame) string {
	if gc.GetPhase() == domain.BigBenPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(gc.GetActionLog())
}
