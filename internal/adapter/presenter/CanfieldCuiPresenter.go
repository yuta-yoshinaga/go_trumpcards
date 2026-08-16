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

// canfieldColumnStr returns the display string for a Canfield tableau column.
func canfieldColumnStr(colCards []*domain.CanfieldTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// CanfieldCuiPresenter renders the Canfield Solitaire CUI view.
type CanfieldCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *CanfieldCuiPresenter) Output(c interfaces.CanfieldGame, lastErr error) string {
	return buildCuiOutput(i18n.T("canfield.helpTitle"), func(b *strings.Builder) {
		// Base rank
		b.WriteString(i18n.Tf("canfield.baseRank",
			"rank", strconv.Itoa(c.GetBaseRank())) + "\n")

		// Foundation
		b.WriteString(i18n.T("canfield.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.CanfieldFoundationCnt {
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

		// Reserve / stock / waste
		reserve := c.GetReserve()
		if len(reserve) > 0 {
			b.WriteString(i18n.Tf("canfield.reserveLine",
				"count", strconv.Itoa(len(reserve)),
				"card", cuiCardStr(reserve[len(reserve)-1])))
		} else {
			b.WriteString(i18n.T("canfield.reserveEmpty"))
		}
		b.WriteString("\n")
		b.WriteString(i18n.Tf("canfield.stockLine",
			"count", strconv.Itoa(c.GetStockCount())))
		waste := c.GetWaste()
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("canfield.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("canfield.wasteEmpty"))
		}
		b.WriteString("\n----------\n")

		// Tableau
		tableau := c.GetTableau()
		for col := range domain.CanfieldTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("canfield.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(canfieldColumnStr(colCards))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.CanfieldPhasePlaying:
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) +
				cuiSolitaireUndoHint(c.CanUndo()) + "\n")
		case domain.CanfieldPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.CanfieldPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Canfield hint.
func (p *CanfieldCuiPresenter) HintOutput(c interfaces.CanfieldGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "tableau":
		from = i18n.Tf("canfield.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	case "reserve":
		from = i18n.T("canfield.hintFromReserve")
	default:
		from = i18n.T("canfield.hintFromWaste")
	}
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("canfield.hintToFoundation")
	} else {
		to = i18n.Tf("canfield.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("canfield.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CanfieldCuiPresenter) ActionLogOutput(c interfaces.CanfieldGame) string {
	if c.GetPhase() == domain.CanfieldPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
