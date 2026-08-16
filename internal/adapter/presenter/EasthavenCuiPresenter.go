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

// EasthavenCuiPresenter renders the Easthaven Solitaire CUI view.
type EasthavenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *EasthavenCuiPresenter) Output(e interfaces.EasthavenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("easthaven.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("easthaven.foundationHeader"))
		foundation := e.GetFoundation()
		for i := range domain.EasthavenFoundationCnt {
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

		// Stock
		b.WriteString(i18n.Tf("easthaven.stockLine", "count", strconv.Itoa(e.GetStockCount())))
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := e.GetTableau()
		for col := range domain.EasthavenTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("easthaven.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(klondikeColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch e.GetPhase() {
		case domain.EasthavenPhasePlaying:
			if e.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := e.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(e.GetMoveCount())) + "\n")
		case domain.EasthavenPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(e.GetMoveCount())) + "\n")
		case domain.EasthavenPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Easthaven hint.
func (p *EasthavenCuiPresenter) HintOutput(e interfaces.EasthavenGame) string {
	hint := e.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("easthaven.hintFrom",
		"col", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("easthaven.hintToFoundation")
	} else {
		to = i18n.Tf("easthaven.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("easthaven.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *EasthavenCuiPresenter) ActionLogOutput(e interfaces.EasthavenGame) string {
	if e.GetPhase() == domain.EasthavenPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(e.GetActionLog())
}
