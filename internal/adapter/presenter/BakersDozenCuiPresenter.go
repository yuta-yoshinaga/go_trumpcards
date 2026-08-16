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

// bakersDozenColumnStr returns the display string for a BakersDozen tableau column.
func bakersDozenColumnStr(colCards []*domain.BakersDozenTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// BakersDozenCuiPresenter renders the Baker's Dozen Solitaire CUI view.
type BakersDozenCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *BakersDozenCuiPresenter) Output(bd interfaces.BakersDozenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bakersdozen.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("bakersdozen.foundationHeader"))
		foundation := bd.GetFoundation()
		for i := range domain.BakersDozenFoundationCnt {
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
		tableau := bd.GetTableau()
		for col := range domain.BakersDozenTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("bakersdozen.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(bakersDozenColumnStr(colCards))
				// A column down to its last card is at risk of emptying, and empty
				// columns can never be refilled — flag it so the risk is visible.
				if len(colCards) == 1 {
					b.WriteString(" " + color.Yellow(i18n.T("bakersdozen.oneCardWarning")))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch bd.GetPhase() {
		case domain.BakersDozenPhasePlaying:
			if bd.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := bd.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.T("bakersdozen.emptyColNote") + "\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(bd.GetMoveCount())) + "\n")
		case domain.BakersDozenPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(bd.GetMoveCount())) + "\n")
		case domain.BakersDozenPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Baker's Dozen hint.
func (p *BakersDozenCuiPresenter) HintOutput(bd interfaces.BakersDozenGame) string {
	hint := bd.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("bakersdozen.hintFrom",
		"col", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("bakersdozen.hintToFoundation")
	} else {
		to = i18n.Tf("bakersdozen.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("bakersdozen.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BakersDozenCuiPresenter) ActionLogOutput(bd interfaces.BakersDozenGame) string {
	return actionLogOutputText(bd)
}
