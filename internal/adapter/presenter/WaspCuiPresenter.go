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

// LegalMovesOutput lists the tableau columns onto which the top (last) card of
// the given column may legally move. Empty columns always accept a card in Wasp
// and are flagged as such. Outside the playing phase, for an out-of-range
// column, or when the column has no movable face-up card, an explanatory line
// is returned instead.
func (p *WaspCuiPresenter) LegalMovesOutput(s interfaces.WaspGame, col int) string {
	if s.GetPhase() != domain.WaspPhasePlaying {
		return i18n.T("wasp.legalNotPlaying") + "\n"
	}
	if col < 0 || col >= domain.WaspTableauCnt {
		return i18n.Tf("invalidColumn", "val", strconv.Itoa(col)) + "\n"
	}
	tableau := s.GetTableau()
	fromCards := tableau[col]
	if len(fromCards) == 0 || !fromCards[len(fromCards)-1].FaceUp {
		return i18n.Tf("wasp.legalNoCard", "col", strconv.Itoa(col)) + "\n"
	}
	moving := fromCards[len(fromCards)-1].Card

	var b strings.Builder
	b.WriteString(i18n.Tf("wasp.legalHeader",
		"col", strconv.Itoa(col),
		"card", cuiCardStr(moving)) + "\n")

	found := false
	for idx := range domain.WaspTableauCnt {
		if idx == col {
			continue
		}
		colCards := tableau[idx]
		if len(colCards) == 0 {
			b.WriteString(i18n.Tf("wasp.legalTargetEmpty", "col", strconv.Itoa(idx)) + "\n")
			found = true
			continue
		}
		top := colCards[len(colCards)-1]
		if top.FaceUp && top.Card.GetDesign() == moving.GetDesign() &&
			top.Card.GetValue() == moving.GetValue()+1 {
			b.WriteString(i18n.Tf("wasp.legalTarget",
				"col", strconv.Itoa(idx),
				"card", cuiCardStr(top.Card)) + "\n")
			found = true
		}
	}
	if !found {
		b.WriteString(i18n.T("wasp.legalNone") + "\n")
	}
	return b.String()
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *WaspCuiPresenter) ActionLogOutput(s interfaces.WaspGame) string {
	if s.GetPhase() == domain.WaspPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
