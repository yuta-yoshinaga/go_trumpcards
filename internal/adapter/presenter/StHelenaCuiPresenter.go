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

// sthelenaColumnStr returns the display string for one tableau column.
func stHelenaColumnStr(colCards []*domain.StHelenaTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// stHelenaRowLabel は列が円のどの位置にあるかを返す。上下は片方の段にしか
// 送れず、左右はどちらへも送れる。
func stHelenaRowLabel(col int) string {
	for _, side := range domain.StHelenaSideColumns {
		if col == side {
			return i18n.T("sthelena.columnRowSide")
		}
	}
	if col < domain.StHelenaTopColumnCnt {
		return i18n.T("sthelena.columnRowTop")
	}
	return i18n.T("sthelena.columnRowBottom")
}

// StHelenaCuiPresenter renders the StHelena Solitaire CUI view.
type StHelenaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *StHelenaCuiPresenter) Output(cr interfaces.StHelenaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sthelena.helpTitle"), func(b *strings.Builder) {
		foundation := cr.GetFoundation()

		// Ascending foundations
		b.WriteString(i18n.T("sthelena.foundationAscendingHeader"))
		for i := range domain.StHelenaAscendingFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			fmt.Fprintf(b, "[%d]", i)
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		// Descending foundations
		b.WriteString(i18n.T("sthelena.foundationDescendingHeader"))
		for i := domain.StHelenaAscendingFoundationCnt; i < domain.StHelenaFoundationCnt; i++ {
			if i != domain.StHelenaAscendingFoundationCnt {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			fmt.Fprintf(b, "[%d]", i)
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		b.WriteString(i18n.Tf("sthelena.redealsLine", "count", strconv.Itoa(cr.GetRedealsRemaining())))
		b.WriteString("\n")
		// **どの列がどの段に送れるかを言う。**制限は盤の見た目からは読めないので、
		// 書かないと「なぜ拒まれたのか」が分からないまま手が止まる。
		if cr.RestrictionsActive() {
			b.WriteString(i18n.T("sthelena.restrictionsLine"))
		} else {
			b.WriteString(i18n.T("sthelena.restrictionsLifted"))
		}
		b.WriteString("\n")
		b.WriteString("----------\n")

		tableau := cr.GetTableau()
		for col := range domain.StHelenaTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("sthelena.columnLabel", "col", strconv.Itoa(col)))
			// 制限が効いている間は、その列がどの段に属するかを添える。
			if cr.RestrictionsActive() {
				b.WriteString(stHelenaRowLabel(col))
			}
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(stHelenaColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch cr.GetPhase() {
		case domain.StHelenaPhasePlaying:
			if cr.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := cr.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(cr.GetMoveCount())) + "\n")
		case domain.StHelenaPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(cr.GetMoveCount())) + "\n")
		case domain.StHelenaPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current StHelena hint.
func (p *StHelenaCuiPresenter) HintOutput(cr interfaces.StHelenaGame) string {
	hint := cr.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	if hint.Redeal {
		return i18n.T("sthelena.hintRedeal") + "\n"
	}
	from := i18n.Tf("sthelena.hintFromTableau", "col", strconv.Itoa(hint.FromCol))
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("sthelena.hintToFoundation", "id", strconv.Itoa(hint.ToCol))
	default:
		to = i18n.Tf("sthelena.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("sthelena.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *StHelenaCuiPresenter) ActionLogOutput(cr interfaces.StHelenaGame) string {
	if cr.GetPhase() == domain.StHelenaPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(cr.GetActionLog())
}
