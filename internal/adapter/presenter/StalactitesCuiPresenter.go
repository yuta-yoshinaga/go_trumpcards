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

// stalactitesEmptyCells は空いているフリーセルの数を返す。
func stalactitesEmptyCells(f interfaces.StalactitesGame) int {
	n := 0
	for _, c := range f.GetCells() {
		if c == nil {
			n++
		}
	}
	return n
}

// stalactitesEmptyColumns は空いているタブロー列の数を返す。
func stalactitesEmptyColumns(f interfaces.StalactitesGame) int {
	n := 0
	for _, col := range f.GetTableau() {
		if len(col) == 0 {
			n++
		}
	}
	return n
}

// StalactitesCuiPresenter renders the Stalactites Solitaire CUI view.
type StalactitesCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *StalactitesCuiPresenter) Output(f interfaces.StalactitesGame, lastErr error) string {
	return buildCuiOutput(i18n.T("stalactites.helpTitle"), func(b *strings.Builder) {
		// Free cells
		b.WriteString(i18n.T("stalactites.stalactitesHeader"))
		cells := f.GetCells()
		for i := 0; i < domain.StalactitesCellCnt; i++ {
			if i != 0 {
				b.WriteString(" | ")
			}
			if cells[i] == nil {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(cells[i]))
			}
		}
		b.WriteString("\n")

		// Foundation
		b.WriteString(i18n.T("stalactites.foundationHeader"))
		foundation := f.GetFoundation()
		for i := 0; i < domain.StalactitesFoundationCnt; i++ {
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

		b.WriteString("----------\n")

		// Tableau
		tableau := f.GetTableau()
		for col := 0; col < domain.StalactitesTableauCnt; col++ {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("stalactites.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				for j, c := range colCards {
					fmt.Fprintf(b, " [%d]%s", j, cuiCardStr(c))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch f.GetPhase() {
		case domain.StalactitesPhasePlaying:
			if f.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := f.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			// **何枚まとめて動かせるかは CUI に出ていなかった (#4777)。**Web は
			// fc-supermove-limit で常時出し、上限を超える列には赤いリングまで
			// 付けている。CUI は空きフリーセル数と空き列数から暗算するか、
			// 実際に動かしてエラーになるまで分からなかった。
			b.WriteString(i18n.Tf("stalactites.supermoveLine",
				"limit", strconv.Itoa(f.GetMaxMovableCards()),
				"cells", strconv.Itoa(stalactitesEmptyCells(f)),
				"cols", strconv.Itoa(stalactitesEmptyColumns(f))))
			// **空き列を移動先にすると上限は下がる。**その列自身を経由地に
			// 使えないため。空き列があるときだけ出す。
			if toEmpty := f.GetMaxMovableCardsToEmptyColumn(); toEmpty > 0 {
				b.WriteString(i18n.Tf("stalactites.supermoveToEmpty",
					"limit", strconv.Itoa(toEmpty)))
			}
			b.WriteString("\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(f.GetMoveCount())) + "\n")
		case domain.StalactitesPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(f.GetMoveCount())) + "\n")
		case domain.StalactitesPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Stalactites hint.
func (p *StalactitesCuiPresenter) HintOutput(f interfaces.StalactitesGame) string {
	hint := f.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "tableau":
		from = i18n.Tf("stalactites.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	case "stalactites":
		from = i18n.Tf("stalactites.hintFromStalactites", "col", strconv.Itoa(hint.FromCol))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.T("stalactites.hintToFoundation")
	case "tableau":
		to = i18n.Tf("stalactites.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	case "stalactites":
		to = i18n.Tf("stalactites.hintToStalactites", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("stalactites.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *StalactitesCuiPresenter) ActionLogOutput(f interfaces.StalactitesGame) string {
	if f.GetPhase() == domain.StalactitesPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(f.GetActionLog())
}
