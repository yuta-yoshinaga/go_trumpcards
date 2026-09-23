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

// citadelColumnStr returns the display string for a Citadel tableau column.
func citadelColumnStr(colCards []*domain.CitadelTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// CitadelCuiPresenter renders the Citadel CUI view.
type CitadelCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CitadelCuiPresenter) Output(bc interfaces.CitadelGame, lastErr error) string {
	return buildCuiOutput(i18n.T("citadel.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("citadel.foundationHeader"))
		foundation := bc.GetFoundation()
		for i := range domain.CitadelFoundationCnt {
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
		for col := range domain.CitadelTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("citadel.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(citadelColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch bc.GetPhase() {
		case domain.CitadelPhasePlaying:
			if bc.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, and the command
				// to use, matching the web StalemateEscapeButton.
				if n := bc.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("citadel.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(bc.GetMoveCount())) + "\n")
		case domain.CitadelPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(bc.GetMoveCount())) + "\n")
		case domain.CitadelPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			fnd := bc.GetFoundation()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				cuiCountPileCards(fnd[:]...), domain.CitadelFoundationCnt*domain.CardValueMax)) + "\n")
		}
	})
}

// HintOutput emits the current Citadel hint.
func (p *CitadelCuiPresenter) HintOutput(bc interfaces.CitadelGame) string {
	hint := bc.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("citadel.hintFrom",
		"col", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("citadel.hintToFoundation")
	} else {
		to = i18n.Tf("citadel.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("citadel.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CitadelCuiPresenter) ActionLogOutput(bc interfaces.CitadelGame) string {
	if bc.GetPhase() == domain.CitadelPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(bc.GetActionLog())
}
