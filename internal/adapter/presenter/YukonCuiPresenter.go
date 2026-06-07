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
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(y.GetMoveCount())) + "\n")
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
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("yukon.hintToFoundation")
	} else {
		to = i18n.Tf("yukon.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("yukon.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *YukonCuiPresenter) ActionLogOutput(y interfaces.YukonGame) string {
	if y.GetPhase() == domain.YukonPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(y.GetActionLog())
}
