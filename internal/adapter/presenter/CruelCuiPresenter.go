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

// CruelCuiPresenter renders the Cruel CUI view.
type CruelCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CruelCuiPresenter) Output(c interfaces.CruelGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cruel.helpTitle"), func(b *strings.Builder) {
		// Foundation row, with each pile's size and the total progress toward 52
		// so the player can gauge how close the game is to being won.
		b.WriteString(i18n.T("cruel.foundationHeader"))
		foundation := c.GetFoundation()
		total := 0
		for i := range domain.CruelFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			total += len(pile)
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(i18n.Tf("cruel.foundationPile",
					"card", cuiCardStr(pile[len(pile)-1]),
					"count", strconv.Itoa(len(pile))))
			}
		}
		b.WriteString("\n")
		b.WriteString(i18n.Tf("cruel.foundationProgress", "total", strconv.Itoa(total)) + "\n")

		b.WriteString("----------\n")

		// Tableau (12 columns).
		tableau := c.GetTableau()
		for col := range domain.CruelTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("cruel.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(klondikeColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.CruelPhasePlaying:
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := c.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
				b.WriteString(i18n.T("cruel.shiftHint") + "\n")
				b.WriteString(color.Yellow(i18n.T("cruel.stalemateGiveUp")) + "\n")
			}
			// CanAutoComplete already shares AutoComplete's own check (#5496), so
			// this line cannot promise a move the command will refuse.
			if c.CanAutoComplete() {
				b.WriteString(color.Green(i18n.T("cruel.autoCompleteReady")) + "\n")
			}
			b.WriteString(i18n.T("cruel.opHelp") + "\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.CruelPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.CruelPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Cruel hint.
func (p *CruelCuiPresenter) HintOutput(c interfaces.CruelGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("cruel.hintFrom", "col", strconv.Itoa(hint.FromCol))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("cruel.hintToFoundation")
	} else {
		to = i18n.Tf("cruel.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("cruel.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CruelCuiPresenter) ActionLogOutput(c interfaces.CruelGame) string {
	if c.GetPhase() == domain.CruelPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
