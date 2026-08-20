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

// BakersGameCuiPresenter renders the Baker's Game Solitaire CUI view. Baker's
// Game shares the FreeCell engine (interfaces.FreeCellGame) and differs only in
// the same-suit stacking rule, so this presenter mirrors FreeCellCuiPresenter
// but resolves its labels under the dedicated "bakersgame" i18n namespace.
type BakersGameCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BakersGameCuiPresenter) Output(f interfaces.FreeCellGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bakersgame.helpTitle"), func(b *strings.Builder) {
		// Free cells
		b.WriteString(i18n.T("bakersgame.freeCellHeader"))
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
		b.WriteString(i18n.T("bakersgame.foundationHeader"))
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
			b.WriteString(i18n.Tf("bakersgame.columnLabel", "col", strconv.Itoa(col)))
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
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := f.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			// **同スート限定のベーカーズ・ゲームでは一括移動の上限がより効く** (#5636)。
			// 上限は FreeCell と同じ (1+空きセル)*2^空き列 で、FreeCell 側は #4777 で
			// 既に出している。Web はバッジで常時出しているので、CUI だけが暗算だった。
			b.WriteString(i18n.Tf("bakersgame.supermoveLine",
				"limit", strconv.Itoa(f.GetMaxMovableCards()),
				"cells", strconv.Itoa(freeCellEmptyCells(f)),
				"cols", strconv.Itoa(freeCellEmptyColumns(f))))
			// 空き列自身は経由地に使えないので、そこへ動かすときの上限は下がる。
			if toEmpty := f.GetMaxMovableCardsToEmptyColumn(); toEmpty > 0 {
				b.WriteString(i18n.Tf("bakersgame.supermoveToEmpty",
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

// HintOutput emits the current Baker's Game hint.
func (p *BakersGameCuiPresenter) HintOutput(f interfaces.FreeCellGame) string {
	hint := f.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "tableau":
		from = i18n.Tf("bakersgame.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	case "freecell":
		from = i18n.Tf("bakersgame.hintFromFreeCell", "col", strconv.Itoa(hint.FromCol))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.T("bakersgame.hintToFoundation")
	case "tableau":
		to = i18n.Tf("bakersgame.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	case "freecell":
		to = i18n.Tf("bakersgame.hintToFreeCell", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("bakersgame.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BakersGameCuiPresenter) ActionLogOutput(f interfaces.FreeCellGame) string {
	if f.GetPhase() == domain.FreeCellPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(f.GetActionLog())
}
