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

// fortyThievesColumnStr returns the display string for a FortyThieves tableau column.
func fortyThievesColumnStr(colCards []*domain.FortyThievesTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// FortyThievesCuiPresenter renders the Forty Thieves Solitaire CUI view.
type FortyThievesCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *FortyThievesCuiPresenter) Output(ft interfaces.FortyThievesGame, lastErr error) string {
	return buildCuiOutput(i18n.T("fortythieves.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("fortythieves.foundationHeader"))
		foundation := ft.GetFoundation()
		for i := range domain.FortyThievesFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(i18n.T("fortythieves.cuiEmptyFoundation"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		// Stock + waste
		b.WriteString(i18n.Tf("fortythieves.stockLine",
			"count", strconv.Itoa(ft.GetStockCount())))
		waste := ft.GetWaste()
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("fortythieves.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("fortythieves.wasteEmpty"))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := ft.GetTableau()
		for col := range domain.FortyThievesTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("fortythieves.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(fortyThievesColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch ft.GetPhase() {
		case domain.FortyThievesPhasePlaying:
			if ft.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := ft.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.T("fortythieves.cuiCommandHint") + "\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(ft.GetMoveCount())) + "\n")
		case domain.FortyThievesPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(ft.GetMoveCount())) + "\n")
		case domain.FortyThievesPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Forty Thieves hint.
func (p *FortyThievesCuiPresenter) HintOutput(ft interfaces.FortyThievesGame) string {
	hint := ft.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	// 盤上に手が無くストックだけ残っている局面。移動の体裁 (「A → B」) に
	// 落とすと列 -1 が漏れるので、専用の文言で言う (#5525)。
	if hint.FromZone == "stock" {
		return i18n.T("fortythieves.hintDraw") + "\n"
	}
	var from string
	if hint.FromZone == "tableau" {
		from = i18n.Tf("fortythieves.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	} else {
		from = i18n.T("fortythieves.hintFromWaste")
	}
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("fortythieves.hintToFoundation")
	} else {
		to = i18n.Tf("fortythieves.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("fortythieves.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FortyThievesCuiPresenter) ActionLogOutput(ft interfaces.FortyThievesGame) string {
	if ft.GetPhase() == domain.FortyThievesPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(ft.GetActionLog())
}
