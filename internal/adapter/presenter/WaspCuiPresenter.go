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

// WaspCuiPresenter renders the Wasp Solitaire CUI view.
type WaspCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *WaspCuiPresenter) Output(s interfaces.WaspGame, lastErr error) string {
	return buildCuiOutput(i18n.T("wasp.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("wasp.header",
			"completed", strconv.Itoa(s.GetCompletedSuits()),
			"total", strconv.Itoa(domain.WaspCompletedSuitsCnt),
			"stock", strconv.Itoa(s.GetStockCount())) + "\n")

		b.WriteString("----------\n")

		tableau := s.GetTableau()
		for col := range domain.WaspTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("wasp.columnLabel", "col", strconv.Itoa(col)))
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
		case domain.WaspPhasePlaying:
			if s.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(s.GetMoveCount())) + "\n")
		case domain.WaspPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(s.GetMoveCount())) + "\n")
		case domain.WaspPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Wasp hint.
func (p *WaspCuiPresenter) HintOutput(s interfaces.WaspGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	if hint.IsDeal() {
		return i18n.T("wasp.hintDeal") + "\n"
	}
	return i18n.Tf("wasp.hintLine",
		"fromCol", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex),
		"toCol", strconv.Itoa(hint.ToCol)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *WaspCuiPresenter) ActionLogOutput(s interfaces.WaspGame) string {
	if s.GetPhase() == domain.WaspPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
