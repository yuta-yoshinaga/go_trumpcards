//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// NarcoticCuiPresenter renders the Narcotic CUI view.
type NarcoticCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (pr *NarcoticCuiPresenter) Output(g interfaces.NarcoticGame, lastErr error) string {
	return buildCuiOutput(i18n.T("narcotic.helpTitle"), func(b *strings.Builder) {
		cols := g.GetColumns()
		for c := range domain.NarcoticColCnt {
			b.WriteString(i18n.Tf("narcotic.columnLabel", "col", strconv.Itoa(c)))
			col := cols[c]
			if len(col) == 0 {
				b.WriteString(i18n.T("narcotic.columnEmpty"))
			} else {
				for i, card := range col {
					if i > 0 {
						b.WriteString(" ")
					}
					if i == len(col)-1 {
						b.WriteString(i18n.Tf("narcotic.topCard", "card", cuiCardStr(card)))
						// Mirror the web UI's enabled/draggable state. **除去は
						// 4枚まとまりなので列ごとの印にはならない** ── 揃っている
						// ときだけ全列に付く。移動は行き先がある列だけ。
						if g.CanRemoveSet() {
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
		b.WriteString(i18n.Tf("narcotic.stockLine", "count", strconv.Itoa(g.GetStockCount())))
		b.WriteString(i18n.Tf("narcotic.discardLine", "count", strconv.Itoa(g.GetDiscardCount())))
		b.WriteString("\n")
		b.WriteString(i18n.T("narcotic.markerLegend") + "\n")
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.NarcoticPhasePlaying:
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
		case domain.NarcoticPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.NarcoticPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Narcotic hint.
func (pr *NarcoticCuiPresenter) HintOutput(g interfaces.NarcoticGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	col := strconv.Itoa(hint.Col)
	// **手だけでなく理由も出す。**Web は同じ場面で hintReason.* の一文を
	// 添えており、CUI だけ「何をするか」しか分からなかった (#5620)。
	switch hint.Type {
	case "remove":
		return i18n.Tf("narcotic.hintRemove", "col", col) + "\n" +
			i18n.Tf("narcotic.hintReasonRemove", "col", col) + "\n"
	case "move":
		return i18n.Tf("narcotic.hintMove", "col", col) + "\n" +
			i18n.Tf("narcotic.hintReasonMove", "col", col) + "\n"
	case "draw":
		return i18n.T("narcotic.hintDraw") + "\n" +
			i18n.T("narcotic.hintReasonDraw") + "\n"
	case "redeal":
		return i18n.T("narcotic.hintRedeal") + "\n" +
			i18n.T("narcotic.hintReasonRedeal") + "\n"
	default:
		return i18n.T("narcotic.hintUnknown") + "\n"
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *NarcoticCuiPresenter) ActionLogOutput(g interfaces.NarcoticGame) string {
	if g.GetPhase() == domain.NarcoticPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
