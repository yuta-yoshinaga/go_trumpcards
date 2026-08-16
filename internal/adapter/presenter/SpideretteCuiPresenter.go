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

// spideretteColumnStr returns the display string for a Spiderette tableau column.
func spideretteColumnStr(colCards []*domain.SpideretteTableauCard) string {
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

// SpideretteCuiPresenter renders the Spiderette Solitaire CUI view.
type SpideretteCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SpideretteCuiPresenter) Output(s interfaces.SpideretteGame, lastErr error) string {
	return buildCuiOutput(i18n.T("spiderette.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("spiderette.header",
			"completed", strconv.Itoa(s.GetCompletedSuits()),
			"total", strconv.Itoa(domain.SpideretteFoundationCnt),
			"stock", strconv.Itoa(s.GetStockCount()),
			"score", strconv.Itoa(s.GetScore())))
		// **生の残り枚数だけでは「あと何回配れるか」が分からない (#4798)。**
		// 1回の配布は最大7枚で、端数の最終配布も1回として数える。Web は同じ
		// 切り上げをバッジに出しているのに、CUI は暗算を強いていた。
		b.WriteString(i18n.Tf("spiderette.dealsRemaining",
			"count", strconv.Itoa(s.GetDealsRemaining())) + "\n")

		b.WriteString("----------\n")

		tableau := s.GetTableau()
		for col := range domain.SpideretteTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("spiderette.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(spideretteColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch s.GetPhase() {
		case domain.SpiderettePhasePlaying:
			if s.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := s.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(s.GetMoveCount())) + "\n")
		case domain.SpiderettePhaseGameClear:
			b.WriteString(color.Green(i18n.Tf("spiderette.gameClearLine",
				"moves", strconv.Itoa(s.GetMoveCount()),
				"score", strconv.Itoa(s.GetScore()))) + "\n")
		case domain.SpiderettePhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Spiderette hint.
func (p *SpideretteCuiPresenter) HintOutput(s interfaces.SpideretteGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	return i18n.Tf("spiderette.hintLine",
		"fromCol", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex),
		"toCol", strconv.Itoa(hint.ToCol)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SpideretteCuiPresenter) ActionLogOutput(s interfaces.SpideretteGame) string {
	return actionLogOutputText(s)
}
