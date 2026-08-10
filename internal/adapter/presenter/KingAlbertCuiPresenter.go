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

// kingAlbertColumnStr returns the display string for a King Albert tableau column.
func kingAlbertColumnStr(colCards []*domain.KingAlbertTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// KingAlbertCuiPresenter renders the King Albert CUI view.
type KingAlbertCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KingAlbertCuiPresenter) Output(bc interfaces.KingAlbertGame, lastErr error) string {
	return buildCuiOutput(i18n.T("kingalbert.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("kingalbert.foundationHeader"))
		foundation := bc.GetFoundation()
		for i := range domain.KingAlbertFoundationCnt {
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

		// Reserve (Lawrence)
		b.WriteString(i18n.T("kingalbert.reserveHeader"))
		reserve := bc.GetReserve()
		for i := range reserve {
			if i != 0 {
				b.WriteString(" ")
			}
			fmt.Fprintf(b, "[r%d]", i)
			if reserve[i] == nil {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(reserve[i]))
			}
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := bc.GetTableau()
		for col := range domain.KingAlbertTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("kingalbert.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(kingAlbertColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch bc.GetPhase() {
		case domain.KingAlbertPhasePlaying:
			if bc.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Point the player at the concrete escape (how many undos are
				// needed), matching the web StalemateEscapeButton.
				//
				// 0 以下は「戻れる局面が無い」。そのまま出すと「undo を -1 回
				// 実行してください」になる (#5052)。
				if n := bc.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("kingalbert.stalemateEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(bc.GetMoveCount())) + "\n")
		case domain.KingAlbertPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(bc.GetMoveCount())) + "\n")
		case domain.KingAlbertPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current King Albert hint.
func (p *KingAlbertCuiPresenter) HintOutput(bc interfaces.KingAlbertGame) string {
	hint := bc.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	if hint.FromZone == "reserve" {
		from = i18n.Tf("kingalbert.hintFromReserve", "idx", strconv.Itoa(hint.FromCol))
	} else {
		from = i18n.Tf("kingalbert.hintFrom",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	}
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("kingalbert.hintToFoundation")
	} else {
		to = i18n.Tf("kingalbert.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("kingalbert.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KingAlbertCuiPresenter) ActionLogOutput(bc interfaces.KingAlbertGame) string {
	if bc.GetPhase() == domain.KingAlbertPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(bc.GetActionLog())
}
