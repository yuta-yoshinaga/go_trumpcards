//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// AccordionCuiPresenter renders the Accordion CUI view.
type AccordionCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *AccordionCuiPresenter) Output(a interfaces.AccordionGame, lastErr error) string {
	return buildCuiOutput(i18n.T("accordion.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("accordion.header",
			"count", strconv.Itoa(a.GetPileCount())) + "\n")
		b.WriteString("----------\n")

		piles := a.GetPiles()
		for i, pile := range piles {
			if len(pile) == 0 {
				continue
			}
			top := pile[len(pile)-1]
			if len(pile) == 1 {
				b.WriteString(i18n.Tf("accordion.pileSingle",
					"idx", strconv.Itoa(i),
					"card", cuiCardStr(top)))
			} else {
				b.WriteString(i18n.Tf("accordion.pileStack",
					"idx", strconv.Itoa(i),
					"card", cuiCardStr(top),
					"count", strconv.Itoa(len(pile)-1)))
			}
			if (i+1)%8 == 0 {
				b.WriteString("\n")
			}
		}
		b.WriteString("\n----------\n")

		cuiErrorBlock(b, lastErr)

		switch a.GetPhase() {
		case domain.AccordionPhasePlaying:
			if a.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := a.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(a.GetMoveCount())) +
				cuiSolitaireUndoHint(a.CanUndo()) + "\n")
		case domain.AccordionPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(a.GetMoveCount())) + "\n")
		case domain.AccordionPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Accordion hint.
func (p *AccordionCuiPresenter) HintOutput(a interfaces.AccordionGame) string {
	hint := a.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	return i18n.Tf("accordion.hintLine",
		"from", strconv.Itoa(hint.FromIdx),
		"to", strconv.Itoa(hint.ToIdx)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *AccordionCuiPresenter) ActionLogOutput(a interfaces.AccordionGame) string {
	if a.GetPhase() == domain.AccordionPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(a.GetActionLog())
}
