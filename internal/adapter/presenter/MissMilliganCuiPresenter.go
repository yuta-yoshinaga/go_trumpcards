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

// missMilliganColumnStr returns the display string for one tableau column.
func missMilliganColumnStr(colCards []*domain.MissMilliganTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// MissMilliganCuiPresenter renders the Miss Milligan CUI view.
type MissMilliganCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MissMilliganCuiPresenter) Output(mm interfaces.MissMilliganGame, lastErr error) string {
	return buildCuiOutput(i18n.T("missmilligan.helpTitle"), func(b *strings.Builder) {
		// 基礎札
		b.WriteString(i18n.T("missmilligan.foundationHeader"))
		foundation := mm.GetFoundation()
		for i := range domain.MissMilliganFoundationCnt {
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

		// 山札と保持中の札。保持中は他の操作がほぼ塞がるので目立たせる。
		b.WriteString(i18n.Tf("missmilligan.stockLine", "count", strconv.Itoa(mm.GetStockCount())))
		waived := mm.GetWaived()
		if len(waived) > 0 {
			parts := make([]string, len(waived))
			for i, c := range waived {
				parts[i] = cuiCardStr(c)
			}
			b.WriteString(" " + color.Yellow(i18n.Tf("missmilligan.waivedLine",
				"cards", strings.Join(parts, " "))))
		} else if mm.CanWaive() {
			b.WriteString(" " + i18n.T("missmilligan.waiveAvailable"))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// タブロー
		tableau := mm.GetTableau()
		for col := range domain.MissMilliganTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("missmilligan.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(missMilliganColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch mm.GetPhase() {
		case domain.MissMilliganPhasePlaying:
			if mm.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, and the command
				// to use, matching the web StalemateEscapeButton.
				if n := mm.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("missmilligan.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(mm.GetMoveCount())) + "\n")
		case domain.MissMilliganPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(mm.GetMoveCount())) + "\n")
		case domain.MissMilliganPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Miss Milligan hint.
func (p *MissMilliganCuiPresenter) HintOutput(mm interfaces.MissMilliganGame) string {
	hint := mm.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "waived":
		from = i18n.T("missmilligan.hintFromWaived")
	case "stock":
		from = i18n.T("missmilligan.hintFromStock")
	default:
		from = i18n.Tf("missmilligan.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	}
	var to string
	switch {
	case hint.ToZone == "foundation":
		to = i18n.Tf("missmilligan.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	case hint.ToIdx < 0:
		// 山札からの配り足しは特定の列を指さない。
		to = i18n.T("missmilligan.hintToDeal")
	default:
		to = i18n.Tf("missmilligan.hintToTableau", "col", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("missmilligan.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MissMilliganCuiPresenter) ActionLogOutput(mm interfaces.MissMilliganGame) string {
	if mm.GetPhase() == domain.MissMilliganPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(mm.GetActionLog())
}
