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

// EightOffCuiPresenter renders the Eight Off Solitaire CUI view.
type EightOffCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *EightOffCuiPresenter) Output(e interfaces.EightOffGame, lastErr error) string {
	return buildCuiOutput(i18n.T("eightoff.helpTitle"), func(b *strings.Builder) {
		// Free cells (8 cells)
		b.WriteString(i18n.T("eightoff.freeCellHeader"))
		freeCells := e.GetFreeCells()
		for i := 0; i < domain.EightOffCellCnt; i++ {
			if i != 0 {
				b.WriteString(" | ")
			}
			if freeCells[i] == nil {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(freeCells[i]))
			}
		}
		b.WriteString("\n")

		// Foundation
		b.WriteString(i18n.T("eightoff.foundationHeader"))
		foundation := e.GetFoundation()
		for i := 0; i < domain.EightOffFoundationCnt; i++ {
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
		tableau := e.GetTableau()
		for col := 0; col < domain.EightOffTableauCnt; col++ {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("eightoff.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				for j, c := range colCards {
					fmt.Fprintf(b, " [%d]%s", j, cuiCardStr(c))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch e.GetPhase() {
		case domain.EightOffPhasePlaying:
			if e.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			// Show how many cards can be moved as one stack — (1 + empty free
			// cells) * 2^(empty columns), the same formula the web UI uses — so the
			// human isn't surprised by a rejected multi-card move.
			emptyCells := 0
			for i := 0; i < domain.EightOffCellCnt; i++ {
				if freeCells[i] == nil {
					emptyCells++
				}
			}
			emptyCols := 0
			for col := 0; col < domain.EightOffTableauCnt; col++ {
				if len(tableau[col]) == 0 {
					emptyCols++
				}
			}
			b.WriteString(i18n.Tf("eightoff.supermoveLine",
				"limit", strconv.Itoa((1+emptyCells)<<emptyCols),
				"cells", strconv.Itoa(emptyCells),
				"cols", strconv.Itoa(emptyCols)) + "\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(e.GetMoveCount())) + "\n")
		case domain.EightOffPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(e.GetMoveCount())) + "\n")
		case domain.EightOffPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Eight Off hint.
func (p *EightOffCuiPresenter) HintOutput(e interfaces.EightOffGame) string {
	hint := e.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "tableau":
		from = i18n.Tf("eightoff.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	case "freecell":
		from = i18n.Tf("eightoff.hintFromFreeCell", "col", strconv.Itoa(hint.FromCol))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.T("eightoff.hintToFoundation")
	case "tableau":
		to = i18n.Tf("eightoff.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	case "freecell":
		to = i18n.Tf("eightoff.hintToFreeCell", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("eightoff.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *EightOffCuiPresenter) ActionLogOutput(e interfaces.EightOffGame) string {
	if e.GetPhase() == domain.EightOffPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(e.GetActionLog())
}
