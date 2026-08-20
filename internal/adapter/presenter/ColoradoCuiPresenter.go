//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// coloradoPileStr returns the display string for one tableau pile.
//
// **動かせるのは一番上の 1 枚だけ。**`m t <山>` は山番号しか取らないのに、
// 以前は全カードに [0][1][2]… と添字を振っていて、埋まった札まで番号で
// 指定できるように見せていた (#5739)。番号を振るのをやめ、一番上だけ
// 印を付ける。
func coloradoPileStr(pile []*domain.Card) string {
	parts := make([]string, len(pile))
	for j, card := range pile {
		if j == len(pile)-1 {
			parts[j] = " " + i18n.Tf("colorado.pileTop", "card", cuiCardStr(card))
			continue
		}
		parts[j] = " " + i18n.Tf("colorado.pileBuried", "card", cuiCardStr(card))
	}
	return strings.Join(parts, " ")
}

// coloradoDirMark returns the arrow that tells an ascending foundation from a
// descending one.
func coloradoDirMark(ascending bool) string {
	if ascending {
		return "\u2191"
	}
	return "\u2193"
}

// ColoradoCuiPresenter renders the Colorado CUI view.
type ColoradoCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ColoradoCuiPresenter) Output(c interfaces.ColoradoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("colorado.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.T("colorado.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.ColoradoFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			// 8 つのうちどれが A→K でどれが K→A かは、表示しないと盤面から読めない。
			b.WriteString(coloradoDirMark(c.IsAscendingFoundation(i)))
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		b.WriteString(i18n.Tf("colorado.stockLine", "count", strconv.Itoa(c.GetStockCount())))
		waste := c.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("colorado.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("colorado.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		b.WriteString(i18n.T("colorado.pileTopNote") + "\n")

		tableau := c.GetTableau()
		for pile := range domain.ColoradoTableauCnt {
			cards := tableau[pile]
			b.WriteString(i18n.Tf("colorado.pileLabel", "pile", strconv.Itoa(pile)))
			if len(cards) == 0 {
				// 空き山は山札からも捨て札からも埋められる。
				b.WriteString(" " + i18n.T("colorado.emptyPile"))
			} else {
				b.WriteString(coloradoPileStr(cards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.ColoradoPhasePlaying:
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := c.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("colorado.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.ColoradoPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.ColoradoPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Colorado hint.
func (p *ColoradoCuiPresenter) HintOutput(c interfaces.ColoradoGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "waste":
		from = i18n.T("colorado.hintFromWaste")
	case "stock":
		from = i18n.T("colorado.hintFromStock")
	default:
		from = i18n.Tf("colorado.hintFromTableau", "pile", strconv.Itoa(hint.FromIdx))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("colorado.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	case "waste":
		to = i18n.T("colorado.hintToWaste")
	default:
		to = i18n.Tf("colorado.hintToTableau", "pile", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("colorado.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ColoradoCuiPresenter) ActionLogOutput(c interfaces.ColoradoGame) string {
	if c.GetPhase() == domain.ColoradoPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
