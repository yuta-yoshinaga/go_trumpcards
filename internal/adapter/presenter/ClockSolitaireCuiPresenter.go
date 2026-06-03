//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ClockSolitaireCuiPresenter renders the Clock Solitaire CUI view.
type ClockSolitaireCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (pr *ClockSolitaireCuiPresenter) Output(g interfaces.ClockSolitaireGame, lastErr error) string {
	return buildCuiOutput(i18n.T("clocksolitaire.helpTitle"), func(b *strings.Builder) {
		piles := g.GetPiles()
		fuc := g.GetFaceUpCount()

		// Clock positions: 12, 1, 2, ..., 11 in display order.
		displayOrder := []int{11, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		labels := []string{"12", " 1", " 2", " 3", " 4", " 5", " 6", " 7", " 8", " 9", "10", "11"}
		total := strconv.Itoa(domain.ClockSolitaireCardsPerPile)

		for i, pileIdx := range displayOrder {
			pile := piles[pileIdx]
			b.WriteString(i18n.Tf("clocksolitaire.hourLabel", "label", labels[i]) + " ")
			for _, pc := range pile {
				if pc.FaceUp {
					b.WriteString(cuiCardStr(pc.Card) + " ")
				} else {
					b.WriteString("?? ")
				}
			}
			b.WriteString(i18n.Tf("clocksolitaire.pileFootnote",
				"up", strconv.Itoa(fuc[pileIdx]),
				"total", total) + "\n")
		}

		// Center pile (kings)
		b.WriteString("----------\n")
		centerPile := piles[domain.ClockSolitaireKingPileIdx]
		b.WriteString(i18n.T("clocksolitaire.centerLabel") + " ")
		for _, pc := range centerPile {
			if pc.FaceUp {
				b.WriteString(cuiCardStr(pc.Card) + " ")
			} else {
				b.WriteString("?? ")
			}
		}
		b.WriteString(i18n.Tf("clocksolitaire.pileFootnote",
			"up", strconv.Itoa(fuc[domain.ClockSolitaireKingPileIdx]),
			"total", total) + "\n")

		b.WriteString("----------\n")

		// Current card (hand)
		if cc := g.GetCurrentCard(); cc != nil {
			b.WriteString(i18n.Tf("clocksolitaire.currentCard", "card", cuiCardStr(cc)) + "\n")
		}

		cuiErrorBlock(b, lastErr)

		count := strconv.Itoa(g.GetStepCount())
		switch g.GetPhase() {
		case domain.ClockSolitairePhasePlaying:
			b.WriteString(i18n.Tf("clocksolitaire.stepLine", "count", count) + "\n")
		case domain.ClockSolitairePhaseGameClear:
			b.WriteString(i18n.Tf("clocksolitaire.gameClearLine", "count", count) + "\n")
		case domain.ClockSolitairePhaseGameOver:
			b.WriteString(i18n.Tf("clocksolitaire.gameOverLine", "count", count) + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *ClockSolitaireCuiPresenter) ActionLogOutput(g interfaces.ClockSolitaireGame) string {
	return buildCuiOutput(i18n.T("clocksolitaire.actionLogTitle"), func(b *strings.Builder) {
		for _, entry := range g.GetActionLog() {
			b.WriteString(i18n.Tf("clocksolitaire.actionLogEntry",
				"turn", strconv.Itoa(entry.TurnNumber),
				"detail", entry.Detail) + "\n")
		}
	})
}
