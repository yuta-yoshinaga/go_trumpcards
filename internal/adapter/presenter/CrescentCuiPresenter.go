//go:build !js || !wasm || solo

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

// crescentColumnStr returns the display string for one tableau column.
func crescentColumnStr(colCards []*domain.CrescentTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// CrescentCuiPresenter renders the Crescent Solitaire CUI view.
type CrescentCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CrescentCuiPresenter) Output(cr interfaces.CrescentGame, lastErr error) string {
	return buildCuiOutput(i18n.T("crescent.helpTitle"), func(b *strings.Builder) {
		foundation := cr.GetFoundation()

		// Ascending foundations
		b.WriteString(i18n.T("crescent.foundationAscendingHeader"))
		for i := range domain.CrescentAscendingFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			fmt.Fprintf(b, "[%d]", i)
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		// Descending foundations
		b.WriteString(i18n.T("crescent.foundationDescendingHeader"))
		for i := domain.CrescentAscendingFoundationCnt; i < domain.CrescentFoundationCnt; i++ {
			if i != domain.CrescentAscendingFoundationCnt {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			fmt.Fprintf(b, "[%d]", i)
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		b.WriteString(i18n.Tf("crescent.redealsLine", "count", strconv.Itoa(cr.GetRedealsRemaining())))
		b.WriteString("\n")
		b.WriteString("----------\n")

		tableau := cr.GetTableau()
		for col := range domain.CrescentTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("crescent.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(crescentColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch cr.GetPhase() {
		case domain.CrescentPhasePlaying:
			if cr.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := cr.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(cr.GetMoveCount())) + "\n")
		case domain.CrescentPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(cr.GetMoveCount())) + "\n")
		case domain.CrescentPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Crescent hint.
func (p *CrescentCuiPresenter) HintOutput(cr interfaces.CrescentGame) string {
	hint := cr.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	if hint.Redeal {
		return i18n.T("crescent.hintRedeal") + "\n"
	}
	from := i18n.Tf("crescent.hintFromTableau", "col", strconv.Itoa(hint.FromCol))
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("crescent.hintToFoundation", "id", strconv.Itoa(hint.ToCol))
	default:
		to = i18n.Tf("crescent.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("crescent.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CrescentCuiPresenter) ActionLogOutput(cr interfaces.CrescentGame) string {
	if cr.GetPhase() == domain.CrescentPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(cr.GetActionLog())
}
