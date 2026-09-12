//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// MatrimonyCuiPresenter renders the Matrimony CUI view.
type MatrimonyCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MatrimonyCuiPresenter) Output(c interfaces.MatrimonyGame, lastErr error) string {
	return buildCuiOutput(i18n.T("matrimony.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.T("matrimony.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.MatrimonyFoundationCnt {
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

		b.WriteString(i18n.Tf("matrimony.stockLine", "count", strconv.Itoa(c.GetStockCount())))
		waste := c.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("matrimony.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("matrimony.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// タブローは 1 枠 1 枚。空き枠は山札か捨て札から補充される。
		tableau := c.GetTableau()
		for slot := range domain.MatrimonyTableauCnt {
			b.WriteString(i18n.Tf("matrimony.slotLabel", "slot", strconv.Itoa(slot)))
			if card := tableau[slot]; card == nil {
				b.WriteString(" " + i18n.T("matrimony.emptySlot"))
			} else {
				b.WriteString(" " + cuiCardStr(card))
			}
			if slot%4 == 3 {
				b.WriteString("\n")
			}
		}

		b.WriteString("----------\n")
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.MatrimonyPhasePlaying:
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := c.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("matrimony.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.MatrimonyPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.MatrimonyPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			fnd := c.GetFoundation()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				cuiCountPileCards(fnd[:]...), domain.MatrimonyFoundationGoal)) + "\n")
		}
	})
}

// HintOutput emits the current Matrimony hint.
func (p *MatrimonyCuiPresenter) HintOutput(c interfaces.MatrimonyGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "waste":
		from = i18n.T("matrimony.hintFromWaste")
	case "stock":
		from = i18n.T("matrimony.hintFromStock")
	default:
		from = i18n.Tf("matrimony.hintFromTableau", "pile", strconv.Itoa(hint.FromIdx))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("matrimony.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	case "waste":
		to = i18n.T("matrimony.hintToWaste")
	default:
		to = i18n.Tf("matrimony.hintToTableau", "pile", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("matrimony.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MatrimonyCuiPresenter) ActionLogOutput(c interfaces.MatrimonyGame) string {
	if c.GetPhase() == domain.MatrimonyPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
