//go:build !js || !wasm || extra3

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

// fortyAndEightColumnStr returns the display string for a FortyAndEight tableau column.
func fortyAndEightColumnStr(colCards []*domain.FortyAndEightTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// FortyAndEightCuiPresenter renders the Forty and Eight Solitaire CUI view.
type FortyAndEightCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *FortyAndEightCuiPresenter) Output(ft interfaces.FortyAndEightGame, lastErr error) string {
	return buildCuiOutput(i18n.T("fortyandeight.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("fortyandeight.foundationHeader"))
		foundation := ft.GetFoundation()
		for i := range domain.FortyAndEightFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			// Label each pile with its move-command index (matching the tableau's
			// column labels) so the eight piles are individually identifiable.
			b.WriteString(i18n.Tf("fortyandeight.foundationLabel", "idx", strconv.Itoa(i)))
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		// Stock + waste
		b.WriteString(i18n.Tf("fortyandeight.stockLine",
			"count", strconv.Itoa(ft.GetStockCount())))
		waste := ft.GetWaste()
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("fortyandeight.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("fortyandeight.wasteEmpty"))
		}
		if ft.GetRedealUsed() {
			b.WriteString(i18n.T("fortyandeight.redealUsed"))
		} else {
			b.WriteString(i18n.T("fortyandeight.redealAvailable"))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := ft.GetTableau()
		for col := range domain.FortyAndEightTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("fortyandeight.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(fortyAndEightColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch ft.GetPhase() {
		case domain.FortyAndEightPhasePlaying:
			if ft.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.T("fortyandeight.cuiCommandHint") + "\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(ft.GetMoveCount())) + "\n")
		case domain.FortyAndEightPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(ft.GetMoveCount())) + "\n")
		case domain.FortyAndEightPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Forty and Eight hint.
func (p *FortyAndEightCuiPresenter) HintOutput(ft interfaces.FortyAndEightGame) string {
	hint := ft.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	if hint.FromZone == "tableau" {
		from = i18n.Tf("fortyandeight.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	} else {
		from = i18n.T("fortyandeight.hintFromWaste")
	}
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("fortyandeight.hintToFoundation")
	} else {
		to = i18n.Tf("fortyandeight.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("fortyandeight.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FortyAndEightCuiPresenter) ActionLogOutput(ft interfaces.FortyAndEightGame) string {
	if ft.GetPhase() == domain.FortyAndEightPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(ft.GetActionLog())
}
