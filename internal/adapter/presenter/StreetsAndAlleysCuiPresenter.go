//go:build !js || !wasm || extra

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

// streetsAndAlleysColumnStr returns the display string for a Streets and Alleys tableau column.
func streetsAndAlleysColumnStr(colCards []*domain.StreetsAndAlleysTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// StreetsAndAlleysCuiPresenter renders the Streets and Alleys CUI view.
type StreetsAndAlleysCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *StreetsAndAlleysCuiPresenter) Output(bc interfaces.StreetsAndAlleysGame, lastErr error) string {
	return buildCuiOutput(i18n.T("streetsandalleys.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("streetsandalleys.foundationHeader"))
		foundation := bc.GetFoundation()
		for i := range domain.StreetsAndAlleysFoundationCnt {
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
		for col := range domain.StreetsAndAlleysTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("streetsandalleys.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(streetsAndAlleysColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch bc.GetPhase() {
		case domain.StreetsAndAlleysPhasePlaying:
			if bc.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := bc.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(bc.GetMoveCount())) + "\n")
		case domain.StreetsAndAlleysPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(bc.GetMoveCount())) + "\n")
		case domain.StreetsAndAlleysPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			fnd := bc.GetFoundation()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				cuiCountPileCards(fnd[:]...), domain.StreetsAndAlleysFoundationCnt*domain.CardValueMax)) + "\n")
		}
	})
}

// HintOutput emits the current Streets and Alleys hint.
func (p *StreetsAndAlleysCuiPresenter) HintOutput(bc interfaces.StreetsAndAlleysGame) string {
	hint := bc.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("streetsandalleys.hintFrom",
		"col", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("streetsandalleys.hintToFoundation")
	} else {
		to = i18n.Tf("streetsandalleys.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("streetsandalleys.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *StreetsAndAlleysCuiPresenter) ActionLogOutput(bc interfaces.StreetsAndAlleysGame) string {
	if bc.GetPhase() == domain.StreetsAndAlleysPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(bc.GetActionLog())
}
