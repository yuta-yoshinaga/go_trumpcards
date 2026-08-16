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

// congressPileStr returns the display string for one tableau pile.
func congressPileStr(pile []*domain.Card) string {
	parts := make([]string, len(pile))
	for j, card := range pile {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(card))
	}
	return strings.Join(parts, " ")
}

// CongressCuiPresenter renders the Congress CUI view.
type CongressCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CongressCuiPresenter) Output(c interfaces.CongressGame, lastErr error) string {
	return buildCuiOutput(i18n.T("congress.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.T("congress.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.CongressFoundationCnt {
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

		b.WriteString(i18n.Tf("congress.stockLine", "count", strconv.Itoa(c.GetStockCount())))
		waste := c.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("congress.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("congress.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		tableau := c.GetTableau()
		for pile := range domain.CongressTableauCnt {
			cards := tableau[pile]
			b.WriteString(i18n.Tf("congress.pileLabel", "pile", strconv.Itoa(pile)))
			if len(cards) == 0 {
				// 空き山は山札か捨て札からしか埋められないので、その旨を添える。
				b.WriteString(" " + i18n.T("congress.emptyPile"))
			} else {
				b.WriteString(congressPileStr(cards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.CongressPhasePlaying:
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := c.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("congress.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.CongressPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.CongressPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			fnd := c.GetFoundation()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				cuiCountPileCards(fnd[:]...), domain.CongressTotalCards)) + "\n")
		}
	})
}

// HintOutput emits the current Congress hint.
func (p *CongressCuiPresenter) HintOutput(c interfaces.CongressGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "waste":
		from = i18n.T("congress.hintFromWaste")
	case "stock":
		from = i18n.T("congress.hintFromStock")
	default:
		from = i18n.Tf("congress.hintFromTableau", "pile", strconv.Itoa(hint.FromIdx))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("congress.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	case "waste":
		to = i18n.T("congress.hintToWaste")
	default:
		to = i18n.Tf("congress.hintToTableau", "pile", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("congress.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CongressCuiPresenter) ActionLogOutput(c interfaces.CongressGame) string {
	if c.GetPhase() == domain.CongressPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
