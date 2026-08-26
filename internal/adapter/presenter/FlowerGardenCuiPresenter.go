//go:build !js || !wasm || extra4

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

// flowerGardenColumnStr returns the display string for a Flower Garden tableau fan.
func flowerGardenColumnStr(colCards []*domain.FlowerGardenTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// FlowerGardenCuiPresenter renders the Flower Garden CUI view.
type FlowerGardenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *FlowerGardenCuiPresenter) Output(bc interfaces.FlowerGardenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("flowergarden.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("flowergarden.foundationHeader"))
		foundation := bc.GetFoundation()
		for i := range domain.FlowerGardenFoundationCnt {
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

		// Reserve (the bouquet). Wrap the 16 cards across rows so the labelled
		// list doesn't overflow an 80-column terminal (the tableau wraps too).
		b.WriteString(i18n.T("flowergarden.reserveHeader"))
		reserve := bc.GetReserve()
		const reservePerRow = 8
		for i := range reserve {
			switch {
			case i == 0:
				// first card follows the header directly
			case i%reservePerRow == 0:
				b.WriteString("\n")
			default:
				b.WriteString(" ")
			}
			fmt.Fprintf(b, "[r%d]", i)
			if reserve[i] == nil {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(reserve[i]))
			}
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := bc.GetTableau()
		for col := range domain.FlowerGardenTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("flowergarden.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(flowerGardenColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch bc.GetPhase() {
		case domain.FlowerGardenPhasePlaying:
			if bc.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := bc.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(bc.GetMoveCount())) + "\n")
		case domain.FlowerGardenPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(bc.GetMoveCount())) + "\n")
		case domain.FlowerGardenPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			fnd := bc.GetFoundation()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				cuiCountPileCards(fnd[:]...), domain.FlowerGardenFoundationCnt*domain.CardValueMax)) + "\n")
		}
	})
}

// HintOutput emits the current Flower Garden hint.
func (p *FlowerGardenCuiPresenter) HintOutput(bc interfaces.FlowerGardenGame) string {
	hint := bc.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	if hint.FromZone == "reserve" {
		from = i18n.Tf("flowergarden.hintFromReserve", "idx", strconv.Itoa(hint.FromCol))
	} else {
		from = i18n.Tf("flowergarden.hintFrom",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	}
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("flowergarden.hintToFoundation")
	} else {
		to = i18n.Tf("flowergarden.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("flowergarden.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FlowerGardenCuiPresenter) ActionLogOutput(bc interfaces.FlowerGardenGame) string {
	if bc.GetPhase() == domain.FlowerGardenPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(bc.GetActionLog())
}
