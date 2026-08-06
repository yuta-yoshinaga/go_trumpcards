//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// SultanCuiPresenter renders the Sultan of Turkey Solitaire CUI view.
type SultanCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *SultanCuiPresenter) Output(su interfaces.SultanGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sultan.helpTitle"), func(b *strings.Builder) {
		// Foundation (each pile shows its top card; the base is a King).
		b.WriteString(i18n.T("sultan.foundationHeader"))
		foundation := su.GetFoundation()
		for i := range domain.SultanFoundationCnt {
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

		// Divan (reserve).
		b.WriteString(i18n.T("sultan.divanHeader"))
		divan := su.GetDivan()
		for i, card := range divan {
			b.WriteString(" [" + strconv.Itoa(i) + "]")
			if card == nil {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(card))
			}
		}
		b.WriteString("\n")

		// Stock + waste.
		b.WriteString(i18n.Tf("sultan.stockLine",
			"count", strconv.Itoa(su.GetStockCount())))
		waste := su.GetWaste()
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("sultan.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("sultan.wasteEmpty"))
		}
		b.WriteString(i18n.Tf("sultan.redealLine",
			"count", strconv.Itoa(domain.SultanMaxRedeal-su.GetRedealCount())))
		b.WriteString("\n")

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch su.GetPhase() {
		case domain.SultanPhasePlaying:
			if su.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// **具体的な脱出手数まで出す。**Web は StalemateEscapeButton に
				// undoToEscape を渡していて、CUI だけ汎用メッセージで止まっていた
				// (#4831)。0 以下は「戻れる局面が無い」なので何も出さない —
				// 「undo を 0 回」は指示にならない。
				if n := su.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("sultan.stalemateEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.T("sultan.cuiCommandHint") + "\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(su.GetMoveCount())) + "\n")
		case domain.SultanPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(su.GetMoveCount())) + "\n")
		case domain.SultanPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Sultan hint.
func (p *SultanCuiPresenter) HintOutput(su interfaces.SultanGame) string {
	hint := su.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	if hint.FromZone == "divan" {
		from = i18n.Tf("sultan.hintFromDivan", "idx", strconv.Itoa(hint.FromIdx))
	} else {
		from = i18n.T("sultan.hintFromWaste")
	}
	to := i18n.Tf("sultan.hintToFoundation", "idx", strconv.Itoa(hint.ToFoundation))
	return i18n.Tf("sultan.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SultanCuiPresenter) ActionLogOutput(su interfaces.SultanGame) string {
	if su.GetPhase() == domain.SultanPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(su.GetActionLog())
}
