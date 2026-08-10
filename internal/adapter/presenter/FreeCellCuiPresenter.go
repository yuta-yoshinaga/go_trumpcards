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

// freeCellEmptyCells は空いているフリーセルの数を返す。
func freeCellEmptyCells(f interfaces.FreeCellGame) int {
	n := 0
	for _, c := range f.GetFreeCells() {
		if c == nil {
			n++
		}
	}
	return n
}

// freeCellEmptyColumns は空いているタブロー列の数を返す。
func freeCellEmptyColumns(f interfaces.FreeCellGame) int {
	n := 0
	for _, col := range f.GetTableau() {
		if len(col) == 0 {
			n++
		}
	}
	return n
}

// FreeCellCuiPresenter renders the FreeCell Solitaire CUI view.
type FreeCellCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *FreeCellCuiPresenter) Output(f interfaces.FreeCellGame, lastErr error) string {
	return buildCuiOutput(i18n.T("freecell.helpTitle"), func(b *strings.Builder) {
		// Free cells
		b.WriteString(i18n.T("freecell.freeCellHeader"))
		freeCells := f.GetFreeCells()
		for i := 0; i < domain.FreeCellCellCnt; i++ {
			if i != 0 {
				b.WriteString(" | ")
			}
			if freeCells[i] == nil {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(freeCells[i]))
			}
		}
		b.WriteString("\n")

		// Foundation
		b.WriteString(i18n.T("freecell.foundationHeader"))
		foundation := f.GetFoundation()
		for i := 0; i < domain.FreeCellFoundationCnt; i++ {
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
		for col := 0; col < domain.FreeCellTableauCnt; col++ {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("freecell.columnLabel", "col", strconv.Itoa(col)))
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
		case domain.FreeCellPhasePlaying:
			if f.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			// **何枚まとめて動かせるかは CUI に出ていなかった (#4777)。**Web は
			// fc-supermove-limit で常時出し、上限を超える列には赤いリングまで
			// 付けている。CUI は空きフリーセル数と空き列数から暗算するか、
			// 実際に動かしてエラーになるまで分からなかった。
			b.WriteString(i18n.Tf("freecell.supermoveLine",
				"limit", strconv.Itoa(f.GetMaxMovableCards()),
				"cells", strconv.Itoa(freeCellEmptyCells(f)),
				"cols", strconv.Itoa(freeCellEmptyColumns(f))))
			// **空き列を移動先にすると上限は下がる。**その列自身を経由地に
			// 使えないため。空き列があるときだけ出す。
			if toEmpty := f.GetMaxMovableCardsToEmptyColumn(); toEmpty > 0 {
				b.WriteString(i18n.Tf("freecell.supermoveToEmpty",
					"limit", strconv.Itoa(toEmpty)))
			}
			b.WriteString("\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(f.GetMoveCount())) + "\n")
		case domain.FreeCellPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(f.GetMoveCount())) + "\n")
		case domain.FreeCellPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current FreeCell hint.
func (p *FreeCellCuiPresenter) HintOutput(f interfaces.FreeCellGame) string {
	hint := f.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "tableau":
		from = i18n.Tf("freecell.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	case "freecell":
		from = i18n.Tf("freecell.hintFromFreeCell", "col", strconv.Itoa(hint.FromCol))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.T("freecell.hintToFoundation")
	case "tableau":
		to = i18n.Tf("freecell.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	case "freecell":
		to = i18n.Tf("freecell.hintToFreeCell", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("freecell.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FreeCellCuiPresenter) ActionLogOutput(f interfaces.FreeCellGame) string {
	if f.GetPhase() == domain.FreeCellPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(f.GetActionLog())
}
