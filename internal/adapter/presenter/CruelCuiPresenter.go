package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// CruelCuiPresenter renders the Cruel CUI view.
type CruelCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CruelCuiPresenter) Output(c interfaces.CruelGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cruel.helpTitle"), func(b *strings.Builder) {
		// Foundation row.
		b.WriteString(i18n.T("cruel.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.CruelFoundationCnt {
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

		// Tableau (12 columns).
		tableau := c.GetTableau()
		for col := range domain.CruelTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("cruel.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(klondikeColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.CruelPhasePlaying:
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				b.WriteString(i18n.T("cruel.shiftHint") + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.CruelPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.CruelPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Cruel hint.
func (p *CruelCuiPresenter) HintOutput(c interfaces.CruelGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("cruel.hintFrom", "col", strconv.Itoa(hint.FromCol))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("cruel.hintToFoundation")
	} else {
		to = i18n.Tf("cruel.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("cruel.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CruelCuiPresenter) ActionLogOutput(c interfaces.CruelGame) string {
	if c.GetPhase() == domain.CruelPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
