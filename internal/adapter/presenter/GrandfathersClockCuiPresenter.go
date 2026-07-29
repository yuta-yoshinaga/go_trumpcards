//go:build !js || !wasm || extra2

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

// grandfathersClockColumnStr returns the display string for one tableau column.
func grandfathersClockColumnStr(colCards []*domain.GrandfathersClockTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// GrandfathersClockCuiPresenter renders the Grandfather's Clock CUI view.
type GrandfathersClockCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *GrandfathersClockCuiPresenter) Output(gc interfaces.GrandfathersClockGame, lastErr error) string {
	return buildCuiOutput(i18n.T("grandfathersclock.helpTitle"), func(b *strings.Builder) {
		// 文字盤。時刻・現在の最上段・目標ランク・完成印を 1 行ずつ出す。
		// 12 個を 1 行に詰めると目標ランクが読めない。
		b.WriteString(i18n.T("grandfathersclock.foundationHeader") + "\n")
		foundation := gc.GetFoundation()
		for i := range domain.GrandfathersClockFoundationCnt {
			b.WriteString(i18n.Tf("grandfathersclock.faceLabel", "idx", strconv.Itoa(i)))
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(" " + cuiCardStr(pile[len(pile)-1]))
			}
			// ランクは数字のまま。cuiCardStr が "SPADE 5" と出す以上、ここだけ
			// A/J/Q/K の表記にすると同じ画面で二通りの読み方が混ざる。
			b.WriteString(" " + i18n.Tf("grandfathersclock.faceTarget",
				"rank", strconv.Itoa(domain.GrandfathersClockTargetRank(i))))
			if gc.IsFoundationComplete(i) {
				b.WriteString(" " + color.Green(i18n.T("grandfathersclock.faceComplete")))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// タブロー
		tableau := gc.GetTableau()
		for col := range domain.GrandfathersClockTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("grandfathersclock.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(grandfathersClockColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch gc.GetPhase() {
		case domain.GrandfathersClockPhasePlaying:
			if gc.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, and the command
				// to use, matching the web StalemateEscapeButton.
				if n := gc.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("grandfathersclock.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(gc.GetMoveCount())) + "\n")
		case domain.GrandfathersClockPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(gc.GetMoveCount())) + "\n")
		case domain.GrandfathersClockPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Grandfather's Clock hint.
func (p *GrandfathersClockCuiPresenter) HintOutput(gc interfaces.GrandfathersClockGame) string {
	hint := gc.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("grandfathersclock.hintFrom", "col", strconv.Itoa(hint.FromCol))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.Tf("grandfathersclock.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	} else {
		to = i18n.Tf("grandfathersclock.hintToTableau", "col", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("grandfathersclock.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GrandfathersClockCuiPresenter) ActionLogOutput(gc interfaces.GrandfathersClockGame) string {
	if gc.GetPhase() == domain.GrandfathersClockPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(gc.GetActionLog())
}
