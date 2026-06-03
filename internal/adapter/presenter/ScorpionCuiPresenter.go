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

// ScorpionCuiPresenter renders the Scorpion Solitaire CUI view.
type ScorpionCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *ScorpionCuiPresenter) Output(s interfaces.ScorpionGame, lastErr error) string {
	return buildCuiOutput(i18n.T("scorpion.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("scorpion.header",
			"completed", strconv.Itoa(s.GetCompletedSuits()),
			"total", strconv.Itoa(domain.ScorpionCompletedSuitsCnt),
			"stock", strconv.Itoa(s.GetStockCount())) + "\n")

		b.WriteString("----------\n")

		tableau := s.GetTableau()
		for col := range domain.ScorpionTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("scorpion.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(klondikeColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch s.GetPhase() {
		case domain.ScorpionPhasePlaying:
			if s.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(s.GetMoveCount())) + "\n")
		case domain.ScorpionPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(s.GetMoveCount())) + "\n")
		case domain.ScorpionPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Scorpion hint.
func (p *ScorpionCuiPresenter) HintOutput(s interfaces.ScorpionGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	if hint.IsDeal() {
		return i18n.T("scorpion.hintDeal") + "\n"
	}
	return i18n.Tf("scorpion.hintLine",
		"fromCol", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex),
		"toCol", strconv.Itoa(hint.ToCol)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ScorpionCuiPresenter) ActionLogOutput(s interfaces.ScorpionGame) string {
	if s.GetPhase() == domain.ScorpionPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
