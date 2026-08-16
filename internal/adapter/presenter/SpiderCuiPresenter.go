//go:build !js || !wasm || solo

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

// spiderColumnStr returns the display string for a Spider tableau column.
func spiderColumnStr(colCards []*domain.SpiderTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		if tc.FaceUp {
			parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
		} else {
			parts[j] = fmt.Sprintf(" [%d]??", j)
		}
	}
	return strings.Join(parts, " ")
}

// SpiderCuiPresenter renders the Spider Solitaire CUI view.
type SpiderCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *SpiderCuiPresenter) Output(s interfaces.SpiderGame, lastErr error) string {
	return buildCuiOutput(i18n.T("spider.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("spider.header",
			"completed", strconv.Itoa(s.GetCompletedSuits()),
			"total", strconv.Itoa(domain.SpiderFoundationCnt),
			"stock", strconv.Itoa(s.GetStockCount()),
			"score", strconv.Itoa(s.GetScore())) + "\n")
		// The difficulty (suit count) and remaining deals (a deal lays one card per
		// column, so stock/10) match the web header.
		b.WriteString(i18n.Tf("spider.difficultyLine",
			"suits", strconv.Itoa(int(s.GetDifficulty())),
			"deals", strconv.Itoa(s.GetStockCount()/domain.SpiderTableauCnt)) + "\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := s.GetTableau()
		emptyColumn := false
		for col := range domain.SpiderTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("spider.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				emptyColumn = true
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(spiderColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		// A deal is blocked while any column is empty (and stock remains); warn up
		// front instead of only surfacing the rejection as an error.
		if emptyColumn && s.GetStockCount() > 0 {
			b.WriteString(color.Yellow(i18n.T("spider.dealBlockedEmpty")) + "\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch s.GetPhase() {
		case domain.SpiderPhasePlaying:
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
		case domain.SpiderPhaseGameClear:
			b.WriteString(color.Green(i18n.Tf("spider.gameClearLine",
				"moves", strconv.Itoa(s.GetMoveCount()),
				"score", strconv.Itoa(s.GetScore()))) + "\n")
		case domain.SpiderPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Spider hint.
func (p *SpiderCuiPresenter) HintOutput(s interfaces.SpiderGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	return i18n.Tf("spider.hintLine",
		"fromCol", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex),
		"toCol", strconv.Itoa(hint.ToCol)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SpiderCuiPresenter) ActionLogOutput(s interfaces.SpiderGame) string {
	if s.GetPhase() == domain.SpiderPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
