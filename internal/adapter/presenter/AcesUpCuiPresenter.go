//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// AcesUpCuiPresenter renders the Aces Up CUI view.
type AcesUpCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (pr *AcesUpCuiPresenter) Output(g interfaces.AcesUpGame, lastErr error) string {
	return buildCuiOutput(i18n.T("acesup.helpTitle"), func(b *strings.Builder) {
		cols := g.GetColumns()
		for c := range domain.AcesUpColCnt {
			b.WriteString(i18n.Tf("acesup.columnLabel", "col", strconv.Itoa(c)))
			col := cols[c]
			if len(col) == 0 {
				b.WriteString(i18n.T("acesup.columnEmpty"))
			} else {
				for i, card := range col {
					if i > 0 {
						b.WriteString(" ")
					}
					if i == len(col)-1 {
						b.WriteString(i18n.Tf("acesup.topCard", "card", cuiCardStr(card)))
						// Mirror the web UI's enabled/draggable state: mark a top card
						// removable (*) when a higher same-suit card sits elsewhere, and
						// movable (>) only when an empty column exists to receive it.
						if g.CanRemove(c) {
							b.WriteString(color.Green("*"))
						}
						if g.CanMove(c) {
							b.WriteString(color.Yellow(">"))
						}
					} else {
						b.WriteString(cuiCardStr(card))
					}
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")
		b.WriteString(i18n.Tf("acesup.stockLine", "count", strconv.Itoa(g.GetStockCount())))
		b.WriteString(i18n.Tf("acesup.discardLine", "count", strconv.Itoa(g.GetDiscardCount())))
		b.WriteString("\n")
		b.WriteString(i18n.T("acesup.markerLegend") + "\n")
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.AcesUpPhasePlaying:
			if g.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := g.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(g.GetMoveCount())) +
				cuiSolitaireUndoHint(g.CanUndo()) + "\n")
		case domain.AcesUpPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.AcesUpPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Aces Up hint.
func (pr *AcesUpCuiPresenter) HintOutput(g interfaces.AcesUpGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	col := strconv.Itoa(hint.Col)
	// **手だけでなく理由も出す。**Web は同じ場面で hintReason.* の一文を
	// 添えており、CUI だけ「何をするか」しか分からなかった (#5620)。
	switch hint.Type {
	case "remove":
		return i18n.Tf("acesup.hintRemove", "col", col) + "\n" +
			i18n.Tf("acesup.hintReasonRemove", "col", col) + "\n"
	case "move":
		return i18n.Tf("acesup.hintMove", "col", col) + "\n" +
			i18n.Tf("acesup.hintReasonMove", "col", col) + "\n"
	case "draw":
		return i18n.T("acesup.hintDraw") + "\n" +
			i18n.T("acesup.hintReasonDraw") + "\n"
	default:
		return i18n.T("acesup.hintUnknown") + "\n"
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *AcesUpCuiPresenter) ActionLogOutput(g interfaces.AcesUpGame) string {
	if g.GetPhase() == domain.AcesUpPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
