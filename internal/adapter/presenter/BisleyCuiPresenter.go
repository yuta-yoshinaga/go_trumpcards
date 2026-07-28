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

// bisleyColumnStr returns the display string for a Bisley tableau column.
func bisleyColumnStr(colCards []*domain.BisleyTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// bisleyFoundationRow renders one foundation row (the top card of each suit pile).
func bisleyFoundationRow(b *strings.Builder, header string, piles [domain.BisleyFoundationCnt][]*domain.Card) {
	b.WriteString(header)
	for i := range domain.BisleyFoundationCnt {
		if i != 0 {
			b.WriteString(" | ")
		}
		pile := piles[i]
		if len(pile) == 0 {
			b.WriteString(i18n.T("cuiEmptyCol"))
		} else {
			b.WriteString(cuiCardStr(pile[len(pile)-1]))
		}
	}
	b.WriteString("\n")
}

// BisleyCuiPresenter renders the Bisley CUI view.
type BisleyCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BisleyCuiPresenter) Output(bg interfaces.BisleyGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bisley.helpTitle"), func(b *strings.Builder) {
		// 基礎札（昇順 A→K / 降順 K→A）
		bisleyFoundationRow(b, i18n.T("bisley.aceFoundationHeader"), bg.GetAceFoundations())
		bisleyFoundationRow(b, i18n.T("bisley.kingFoundationHeader"), bg.GetKingFoundations())

		b.WriteString("----------\n")

		// タブロー
		tableau := bg.GetTableau()
		for col := range domain.BisleyTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("bisley.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(bisleyColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch bg.GetPhase() {
		case domain.BisleyPhasePlaying:
			if bg.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, and the command
				// to use, matching the web StalemateEscapeButton.
				if n := bg.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("bisley.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(bg.GetMoveCount())) + "\n")
		case domain.BisleyPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(bg.GetMoveCount())) + "\n")
		case domain.BisleyPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Bisley hint.
func (p *BisleyCuiPresenter) HintOutput(bg interfaces.BisleyGame) string {
	hint := bg.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("bisley.hintFrom", "col", strconv.Itoa(hint.FromCol))
	var to string
	switch hint.ToZone {
	case "ace":
		to = i18n.Tf("bisley.hintToAce", "idx", strconv.Itoa(hint.ToIdx))
	case "king":
		to = i18n.Tf("bisley.hintToKing", "idx", strconv.Itoa(hint.ToIdx))
	default:
		to = i18n.Tf("bisley.hintToTableau", "col", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("bisley.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BisleyCuiPresenter) ActionLogOutput(bg interfaces.BisleyGame) string {
	if bg.GetPhase() == domain.BisleyPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(bg.GetActionLog())
}
