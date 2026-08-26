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

// fortressColumnStr returns the display string for a Fortress tableau column.
func fortressColumnStr(colCards []*domain.FortressTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// FortressCuiPresenter renders the Fortress CUI view.
type FortressCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *FortressCuiPresenter) Output(bc interfaces.FortressGame, lastErr error) string {
	return buildCuiOutput(i18n.T("fortress.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("fortress.foundationHeader"))
		foundation := bc.GetFoundation()
		for i := range domain.FortressFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := bc.GetTableau()
		for col := range domain.FortressTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("fortress.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(fortressColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch bc.GetPhase() {
		case domain.FortressPhasePlaying:
			if bc.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, and the command
				// to use, matching the web StalemateEscapeButton.
				if n := bc.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("fortress.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(bc.GetMoveCount())) + "\n")
		case domain.FortressPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(bc.GetMoveCount())) + "\n")
		case domain.FortressPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			fnd := bc.GetFoundation()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				cuiCountPileCards(fnd[:]...), domain.FortressFoundationCnt*domain.CardValueMax)) + "\n")
		}
	})
}

// HintOutput emits the current Fortress hint.
func (p *FortressCuiPresenter) HintOutput(bc interfaces.FortressGame) string {
	hint := bc.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("fortress.hintFrom",
		"col", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("fortress.hintToFoundation")
	} else {
		to = i18n.Tf("fortress.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("fortress.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FortressCuiPresenter) ActionLogOutput(bc interfaces.FortressGame) string {
	if bc.GetPhase() == domain.FortressPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(bc.GetActionLog())
}
