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

// SeahavenTowersCuiPresenter renders the Seahaven Towers solitaire CUI view.
type SeahavenTowersCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SeahavenTowersCuiPresenter) Output(s interfaces.SeahavenTowersGame, lastErr error) string {
	return buildCuiOutput(i18n.T("seahaventowers.helpTitle"), func(b *strings.Builder) {
		// Reserved cells (towers).
		b.WriteString(i18n.T("seahaventowers.reservedHeader"))
		cells := s.GetFreeCells()
		for i := 0; i < domain.SeahavenTowersCellCnt; i++ {
			if i != 0 {
				b.WriteString(" | ")
			}
			if cells[i] == nil {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(cells[i]))
			}
		}
		b.WriteString("\n")

		// Foundation.
		b.WriteString(i18n.T("seahaventowers.foundationHeader"))
		foundation := s.GetFoundation()
		for i := 0; i < domain.SeahavenTowersFoundationCnt; i++ {
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

		// Tableau.
		tableau := s.GetTableau()
		for col := 0; col < domain.SeahavenTowersTableauCnt; col++ {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("seahaventowers.columnLabel", "col", strconv.Itoa(col)))
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

		switch s.GetPhase() {
		case domain.SeahavenTowersPhasePlaying:
			if s.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			// Show the one-move stack limit (1 + empty reserved cells, the web
			// formula) so the human isn't surprised by a rejected multi-card move.
			emptyReserved := 0
			for i := 0; i < domain.SeahavenTowersCellCnt; i++ {
				if cells[i] == nil {
					emptyReserved++
				}
			}
			b.WriteString(i18n.Tf("seahaventowers.supermoveLine",
				"limit", strconv.Itoa(1+emptyReserved),
				"reserved", strconv.Itoa(emptyReserved)) + "\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(s.GetMoveCount())) + "\n")
		case domain.SeahavenTowersPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(s.GetMoveCount())) + "\n")
		case domain.SeahavenTowersPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Seahaven Towers hint.
func (p *SeahavenTowersCuiPresenter) HintOutput(s interfaces.SeahavenTowersGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "tableau":
		from = i18n.Tf("seahaventowers.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	case "reserved":
		from = i18n.Tf("seahaventowers.hintFromReserved", "col", strconv.Itoa(hint.FromCol))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.T("seahaventowers.hintToFoundation")
	case "tableau":
		to = i18n.Tf("seahaventowers.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	case "reserved":
		to = i18n.Tf("seahaventowers.hintToReserved", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("seahaventowers.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SeahavenTowersCuiPresenter) ActionLogOutput(s interfaces.SeahavenTowersGame) string {
	if s.GetPhase() == domain.SeahavenTowersPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
