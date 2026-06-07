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

// TriPeaksCuiPresenter renders the TriPeaks Solitaire CUI view.
type TriPeaksCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (pr *TriPeaksCuiPresenter) Output(t interfaces.TriPeaksGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tripeaks.helpTitle"), func(b *strings.Builder) {
		layout := t.GetLayout()

		// Tableau (with row indent)
		for row := range domain.TriPeaksRowCnt {
			indent := strings.Repeat("  ", domain.TriPeaksRowCnt-1-row)
			b.WriteString(indent)
			first := true
			for col := range domain.TriPeaksColCnt {
				tc := layout[row][col]
				if tc == nil {
					continue
				}
				if !first {
					b.WriteString("  ")
				}
				first = false
				if tc.Removed {
					b.WriteString("    ")
				} else {
					b.WriteString(i18n.Tf("tripeaks.tableauCard",
						"row", strconv.Itoa(row),
						"col", strconv.Itoa(col),
						"card", cuiCardStr(tc.Card)))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// Stock + waste
		b.WriteString(i18n.Tf("tripeaks.stockLine",
			"count", strconv.Itoa(t.GetStockCount())))
		waste := t.GetWaste()
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("tripeaks.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("tripeaks.wasteEmpty"))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch t.GetPhase() {
		case domain.TriPeaksPhasePlaying:
			if t.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(t.GetMoveCount())) + "\n")
		case domain.TriPeaksPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(t.GetMoveCount())) + "\n")
		case domain.TriPeaksPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current TriPeaks hint.
func (pr *TriPeaksCuiPresenter) HintOutput(t interfaces.TriPeaksGame) string {
	hint := t.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	switch hint.Type {
	case "remove":
		return i18n.Tf("tripeaks.hintRemove",
			"row", strconv.Itoa(hint.Row),
			"col", strconv.Itoa(hint.Col)) + "\n"
	case "draw":
		return i18n.T("tripeaks.hintDraw") + "\n"
	default:
		return i18n.T("tripeaks.hintUnknown") + "\n"
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *TriPeaksCuiPresenter) ActionLogOutput(t interfaces.TriPeaksGame) string {
	if t.GetPhase() == domain.TriPeaksPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(t.GetActionLog())
}
