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

// crazyQuiltSeriesMark tells an Ace-start foundation from a King-start one.
func crazyQuiltSeriesMark(ascending bool) string {
	if ascending {
		return "\u2191"
	}
	return "\u2193"
}

// CrazyQuiltCuiPresenter renders the CrazyQuilt CUI view.
type CrazyQuiltCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CrazyQuiltCuiPresenter) Output(c interfaces.CrazyQuiltGame, lastErr error) string {
	return buildCuiOutput(i18n.T("crazyquilt.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.T("crazyquilt.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.CrazyQuiltFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			// A 始まりか K 始まりかを出さないと、次に何が要るのか読めない。
			b.WriteString(crazyQuiltSeriesMark(c.IsAscendingFoundation(i)))
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		b.WriteString(i18n.Tf("crazyquilt.stockLine", "count", strconv.Itoa(c.GetStockCount())))
		b.WriteString(" " + i18n.Tf("crazyquilt.redealLine", "count", strconv.Itoa(c.GetRedealsLeft())))
		waste := c.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("crazyquilt.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("crazyquilt.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// キルトは 8×8。**取れる札には印を付ける** — 短辺が露出しているかは
		// 向きに依存するので、盤面を見ただけでは分からない。
		quilt := c.GetQuilt()
		for row := range domain.CrazyQuiltGridSize {
			for col := range domain.CrazyQuiltGridSize {
				idx := row*domain.CrazyQuiltGridSize + col
				card := quilt[idx]
				if card == nil {
					b.WriteString(i18n.T("crazyquilt.emptyCell"))
					continue
				}
				mark := " "
				if c.IsAvailable(idx) {
					mark = "*"
				}
				b.WriteString(mark + cuiCardStr(card))
			}
			b.WriteString("\n")
		}
		b.WriteString(i18n.T("crazyquilt.availableLegend") + "\n")

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.CrazyQuiltPhasePlaying:
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := c.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("crazyquilt.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.CrazyQuiltPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.CrazyQuiltPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			fnd := c.GetFoundation()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				cuiCountPileCards(fnd[:]...), domain.CrazyQuiltTotalCards)) + "\n")
		}
	})
}

// HintOutput emits the current CrazyQuilt hint.
func (p *CrazyQuiltCuiPresenter) HintOutput(c interfaces.CrazyQuiltGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "waste":
		from = i18n.T("crazyquilt.hintFromWaste")
	case "stock":
		from = i18n.T("crazyquilt.hintFromStock")
	default:
		from = i18n.Tf("crazyquilt.hintFromQuilt", "cell", strconv.Itoa(hint.FromIdx))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("crazyquilt.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	default:
		// 行き先はキルトにならない（キルトは崩す一方）ので、残りは捨て札だけ。
		to = i18n.T("crazyquilt.hintToWaste")
	}
	return i18n.Tf("crazyquilt.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CrazyQuiltCuiPresenter) ActionLogOutput(c interfaces.CrazyQuiltGame) string {
	if c.GetPhase() == domain.CrazyQuiltPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
