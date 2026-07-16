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

// PenguinCuiPresenter renders the Penguin Solitaire CUI view.
type PenguinCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PenguinCuiPresenter) Output(pg interfaces.PenguinGame, lastErr error) string {
	return buildCuiOutput(i18n.T("penguin.helpTitle"), func(b *strings.Builder) {
		// Free cells (7 cells)
		b.WriteString(i18n.T("penguin.freeCellHeader"))
		freeCells := pg.GetFreeCells()
		for i := 0; i < domain.PenguinCellCnt; i++ {
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
		b.WriteString(i18n.T("penguin.foundationHeader"))
		foundation := pg.GetFoundation()
		for i := 0; i < domain.PenguinFoundationCnt; i++ {
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

		// Base rank — render 1/11/12/13 as A/J/Q/K (matching the web baseRankLabel
		// and cuiCardStr) rather than a raw number.
		b.WriteString(i18n.Tf("penguin.baseRankLabel", "rank", cuiRankLabel(pg.GetBaseRank())))
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := pg.GetTableau()
		for col := 0; col < domain.PenguinTableauCnt; col++ {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("penguin.columnLabel", "col", strconv.Itoa(col)))
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

		switch pg.GetPhase() {
		case domain.PenguinPhasePlaying:
			if pg.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(pg.GetMoveCount())) + "\n")
		case domain.PenguinPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(pg.GetMoveCount())) + "\n")
		case domain.PenguinPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Penguin hint.
func (p *PenguinCuiPresenter) HintOutput(pg interfaces.PenguinGame) string {
	hint := pg.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "tableau":
		from = i18n.Tf("penguin.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	case "freecell":
		from = i18n.Tf("penguin.hintFromFreeCell", "col", strconv.Itoa(hint.FromCol))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.T("penguin.hintToFoundation")
	case "tableau":
		to = i18n.Tf("penguin.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	case "freecell":
		to = i18n.Tf("penguin.hintToFreeCell", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("penguin.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PenguinCuiPresenter) ActionLogOutput(pg interfaces.PenguinGame) string {
	if pg.GetPhase() == domain.PenguinPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(pg.GetActionLog())
}
