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

// RussianSolitaireCuiPresenter renders the Russian Solitaire CUI view.
type RussianSolitaireCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *RussianSolitaireCuiPresenter) Output(r interfaces.RussianSolitaireGame, lastErr error) string {
	return buildCuiOutput(i18n.T("russiansolitaire.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("russiansolitaire.foundationHeader"))
		foundation := r.GetFoundation()
		for i := range domain.RussianSolitaireFoundationCnt {
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
		tableau := r.GetTableau()
		for col := range domain.RussianSolitaireTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("russiansolitaire.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(klondikeColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch r.GetPhase() {
		case domain.RussianSolitairePhasePlaying:
			if r.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := r.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(r.GetMoveCount())) + "\n")
		case domain.RussianSolitairePhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(r.GetMoveCount())) + "\n")
		case domain.RussianSolitairePhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Russian Solitaire hint.
func (p *RussianSolitaireCuiPresenter) HintOutput(r interfaces.RussianSolitaireGame) string {
	hint := r.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("russiansolitaire.hintFrom",
		"col", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("russiansolitaire.hintToFoundation")
	} else {
		to = i18n.Tf("russiansolitaire.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("russiansolitaire.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *RussianSolitaireCuiPresenter) ActionLogOutput(r interfaces.RussianSolitaireGame) string {
	if r.GetPhase() == domain.RussianSolitairePhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(r.GetActionLog())
}
