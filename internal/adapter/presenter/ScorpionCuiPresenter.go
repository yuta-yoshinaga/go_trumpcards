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
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := s.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
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
	line := i18n.Tf("scorpion.hintLine",
		"fromCol", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex),
		"toCol", strconv.Itoa(hint.ToCol))
	// **なぜその手なのかを言う。**GetHint は裏カードを開ける手を先に探しており、
	// 移動先だけ見せてもプレイヤーはその優先順位を学べない (#5544)。
	if hint.ExposesFaceDown {
		line += i18n.T("scorpion.hintExposes")
	}
	return line + "\n"
}

// LegalMovesOutput lists the tableau columns onto which the top (last) card of
// the given column may legally move. In Scorpion an empty column only accepts a
// King, so empty columns are listed only when the moving card is a King.
// Outside the playing phase, for an out-of-range column, or when the column has
// no movable face-up card, an explanatory line is returned instead.
func (p *ScorpionCuiPresenter) LegalMovesOutput(s interfaces.ScorpionGame, col int) string {
	if s.GetPhase() != domain.ScorpionPhasePlaying {
		return i18n.T("scorpion.legalNotPlaying") + "\n"
	}
	if col < 0 || col >= domain.ScorpionTableauCnt {
		return i18n.Tf("invalidColumn", "val", strconv.Itoa(col)) + "\n"
	}
	tableau := s.GetTableau()
	fromCards := tableau[col]
	if len(fromCards) == 0 || !fromCards[len(fromCards)-1].FaceUp {
		return i18n.Tf("scorpion.legalNoCard", "col", strconv.Itoa(col)) + "\n"
	}
	moving := fromCards[len(fromCards)-1].Card
	isKing := moving.GetValue() == domain.CardValueMax

	var b strings.Builder
	b.WriteString(i18n.Tf("scorpion.legalHeader",
		"col", strconv.Itoa(col),
		"card", cuiCardStr(moving)) + "\n")

	found := false
	for idx := range domain.ScorpionTableauCnt {
		if idx == col {
			continue
		}
		colCards := tableau[idx]
		if len(colCards) == 0 {
			if isKing {
				b.WriteString(i18n.Tf("scorpion.legalTargetEmpty", "col", strconv.Itoa(idx)) + "\n")
				found = true
			}
			continue
		}
		top := colCards[len(colCards)-1]
		if top.FaceUp && top.Card.GetDesign() == moving.GetDesign() &&
			top.Card.GetValue() == moving.GetValue()+1 {
			b.WriteString(i18n.Tf("scorpion.legalTarget",
				"col", strconv.Itoa(idx),
				"card", cuiCardStr(top.Card)) + "\n")
			found = true
		}
	}
	if !found {
		b.WriteString(i18n.T("scorpion.legalNone") + "\n")
	}
	return b.String()
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ScorpionCuiPresenter) ActionLogOutput(s interfaces.ScorpionGame) string {
	if s.GetPhase() == domain.ScorpionPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
