//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// YukonCuiPresenter renders the Yukon Solitaire CUI view.
type YukonCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *YukonCuiPresenter) Output(y interfaces.YukonGame, lastErr error) string {
	return buildCuiOutput(i18n.T("yukon.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("yukon.foundationHeader"))
		foundation := y.GetFoundation()
		for i := range domain.YukonFoundationCnt {
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
		tableau := y.GetTableau()
		for col := range domain.YukonTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("yukon.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(klondikeColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch y.GetPhase() {
		case domain.YukonPhasePlaying:
			if y.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := y.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(y.GetMoveCount())) + "\n")
			// 盤面は Klondike とまったく同じ見た目なので、Klondike 経験者は
			// 「揃った並びしか動かせない」と思い込んだままになる。Web 版が専用の
			// ホバープレビューまで用意しているルールを、CUI でも明示する (#4788)。
			b.WriteString(i18n.T("yukon.blockMoveRule") + "\n")
		case domain.YukonPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(y.GetMoveCount())) + "\n")
		case domain.YukonPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Yukon hint.
func (p *YukonCuiPresenter) HintOutput(y interfaces.YukonGame) string {
	hint := y.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("yukon.hintFrom",
		"col", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex))
	var to, confidence string
	if hint.ToZone == "foundation" {
		// Foundation moves never hurt, so they are the highest-priority play.
		to = i18n.T("yukon.hintToFoundation")
		confidence = i18n.T("yukon.hintConfidenceFoundation")
	} else {
		to = i18n.Tf("yukon.hintToTableau", "col", strconv.Itoa(hint.ToCol))
		confidence = i18n.T("yukon.hintConfidenceTableau")
	}
	return i18n.Tf("yukon.hintLine", "from", from, "to", to, "confidence", confidence) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *YukonCuiPresenter) ActionLogOutput(y interfaces.YukonGame) string {
	if y.GetPhase() == domain.YukonPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(y.GetActionLog())
}
