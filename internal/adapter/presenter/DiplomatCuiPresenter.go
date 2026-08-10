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

// diplomatPileStr returns the display string for one tableau pile.
func diplomatPileStr(pile []*domain.Card) string {
	parts := make([]string, len(pile))
	for j, card := range pile {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(card))
	}
	return strings.Join(parts, " ")
}

// DiplomatCuiPresenter renders the Diplomat CUI view.
type DiplomatCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *DiplomatCuiPresenter) Output(c interfaces.DiplomatGame, lastErr error) string {
	return buildCuiOutput(i18n.T("diplomat.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.T("diplomat.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.DiplomatFoundationCnt {
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

		b.WriteString(i18n.Tf("diplomat.stockLine", "count", strconv.Itoa(c.GetStockCount())))
		waste := c.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("diplomat.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("diplomat.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		tableau := c.GetTableau()
		for pile := range domain.DiplomatTableauCnt {
			cards := tableau[pile]
			b.WriteString(i18n.Tf("diplomat.pileLabel", "pile", strconv.Itoa(pile)))
			if len(cards) == 0 {
				// 空き山は山札か捨て札からしか埋められないので、その旨を添える。
				b.WriteString(" " + i18n.T("diplomat.emptyPile"))
			} else {
				b.WriteString(diplomatPileStr(cards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.DiplomatPhasePlaying:
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := c.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("diplomat.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.DiplomatPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.DiplomatPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Diplomat hint.
func (p *DiplomatCuiPresenter) HintOutput(c interfaces.DiplomatGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "waste":
		from = i18n.T("diplomat.hintFromWaste")
	case "stock":
		from = i18n.T("diplomat.hintFromStock")
	default:
		from = i18n.Tf("diplomat.hintFromTableau", "pile", strconv.Itoa(hint.FromIdx))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("diplomat.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	case "waste":
		to = i18n.T("diplomat.hintToWaste")
	default:
		to = i18n.Tf("diplomat.hintToTableau", "pile", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("diplomat.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *DiplomatCuiPresenter) ActionLogOutput(c interfaces.DiplomatGame) string {
	if c.GetPhase() == domain.DiplomatPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
