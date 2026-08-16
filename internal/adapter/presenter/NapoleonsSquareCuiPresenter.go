//go:build !js || !wasm || extra2

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

// napoleonsSquareColumnStr returns the display string for one tableau column.
func napoleonsSquareColumnStr(colCards []*domain.NapoleonsSquareTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// NapoleonsSquareCuiPresenter renders the Napoleon's Square CUI view.
type NapoleonsSquareCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *NapoleonsSquareCuiPresenter) Output(ns interfaces.NapoleonsSquareGame, lastErr error) string {
	return buildCuiOutput(i18n.T("napoleonssquare.helpTitle"), func(b *strings.Builder) {
		// 基礎札（スートごとに 2 つ）
		b.WriteString(i18n.T("napoleonssquare.foundationHeader"))
		foundation := ns.GetFoundation()
		for i := range domain.NapoleonsSquareFoundationCnt {
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

		// 山札とウェイスト
		b.WriteString(i18n.Tf("napoleonssquare.stockLine",
			"count", strconv.Itoa(ns.GetStockCount())))
		waste := ns.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("napoleonssquare.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("napoleonssquare.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// タブロー
		tableau := ns.GetTableau()
		for col := range domain.NapoleonsSquareTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("napoleonssquare.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(napoleonsSquareColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch ns.GetPhase() {
		case domain.NapoleonsSquarePhasePlaying:
			if ns.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, and the command
				// to use, matching the web StalemateEscapeButton.
				if n := ns.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("napoleonssquare.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(ns.GetMoveCount())) + "\n")
		case domain.NapoleonsSquarePhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(ns.GetMoveCount())) + "\n")
		case domain.NapoleonsSquarePhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			fnd := ns.GetFoundation()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				cuiCountPileCards(fnd[:]...), domain.NapoleonsSquareFoundationCnt*domain.CardValueMax)) + "\n")
		}
	})
}

// HintOutput emits the current Napoleon's Square hint.
func (p *NapoleonsSquareCuiPresenter) HintOutput(ns interfaces.NapoleonsSquareGame) string {
	hint := ns.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "waste":
		from = i18n.T("napoleonssquare.hintFromWaste")
	case "stock":
		from = i18n.T("napoleonssquare.hintFromStock")
	default:
		from = i18n.Tf("napoleonssquare.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("napoleonssquare.hintToFoundation", "idx", strconv.Itoa(hint.ToCol))
	case "waste":
		to = i18n.T("napoleonssquare.hintToWaste")
	default:
		to = i18n.Tf("napoleonssquare.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("napoleonssquare.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *NapoleonsSquareCuiPresenter) ActionLogOutput(ns interfaces.NapoleonsSquareGame) string {
	if ns.GetPhase() == domain.NapoleonsSquarePhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(ns.GetActionLog())
}
