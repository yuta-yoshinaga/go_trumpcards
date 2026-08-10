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

// agnesColumnStr returns the display string for an Agnes tableau column.
func agnesColumnStr(colCards []*domain.AgnesTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		if tc.FaceUp {
			parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
		} else {
			parts[j] = fmt.Sprintf(" [%d]??", j)
		}
	}
	return strings.Join(parts, " ")
}

// AgnesCuiPresenter renders the Agnes Sorel Solitaire CUI view.
type AgnesCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *AgnesCuiPresenter) Output(c interfaces.AgnesGame, lastErr error) string {
	return buildCuiOutput(i18n.T("agnes.helpTitle"), func(b *strings.Builder) {
		// Base rank
		b.WriteString(i18n.Tf("agnes.baseRank",
			"rank", strconv.Itoa(c.GetBaseRank())) + "\n")

		// Foundation
		b.WriteString(i18n.T("agnes.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.AgnesFoundationCnt {
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
		b.WriteString(i18n.Tf("agnes.stockLine",
			"count", strconv.Itoa(c.GetStockCount())))
		b.WriteString("\n----------\n")

		// Tableau
		tableau := c.GetTableau()
		for col := range domain.AgnesTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("agnes.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(agnesColumnStr(colCards))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.AgnesPhasePlaying:
			// Web は ag-stalemate-banner で毎レンダー手詰まりを知らせているのに、
			// CUI は手数しか出しておらず、詰んでいても分からなかった (#4830)。
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.AgnesPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.AgnesPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Agnes hint.
func (p *AgnesCuiPresenter) HintOutput(c interfaces.AgnesGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("agnes.hintFromTableau",
		"col", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("agnes.hintToFoundation")
	} else {
		to = i18n.Tf("agnes.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("agnes.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *AgnesCuiPresenter) ActionLogOutput(c interfaces.AgnesGame) string {
	if c.GetPhase() == domain.AgnesPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
