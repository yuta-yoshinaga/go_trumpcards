//go:build !js || !wasm || extra3

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

// terracePileStr returns the display string for one tableau pile.
func terracePileStr(pile []*domain.Card) string {
	parts := make([]string, len(pile))
	for j, card := range pile {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(card))
	}
	return strings.Join(parts, " ")
}

// TerraceCuiPresenter renders the Terrace CUI view.
type TerraceCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TerraceCuiPresenter) Output(t interfaces.TerraceGame, lastErr error) string {
	return buildCuiOutput(i18n.T("terrace.helpTitle"), func(b *strings.Builder) {
		// 開始ランクが未決定のうちは、その 1 手が最重要なので先頭に出す。
		if t.IsAwaitingBaseRank() {
			b.WriteString(color.Yellow(i18n.T("terrace.awaitingBase")) + "\n")
		} else {
			b.WriteString(i18n.Tf("terrace.baseRankLine",
				"rank", strconv.Itoa(t.GetBaseRank())) + "\n")
		}

		b.WriteString(i18n.T("terrace.foundationHeader"))
		foundation := t.GetFoundation()
		for i := range domain.TerraceFoundationCnt {
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

		// テラスは基礎札にしか出せず、補充もされない。残り枚数がそのまま重みになる。
		reserve := t.GetReserve()
		if len(reserve) == 0 {
			b.WriteString(i18n.T("terrace.reserveEmpty") + "\n")
		} else {
			b.WriteString(i18n.Tf("terrace.reserveLine",
				"card", cuiCardStr(reserve[len(reserve)-1]),
				"count", strconv.Itoa(len(reserve))) + "\n")
		}

		b.WriteString(i18n.Tf("terrace.stockLine", "count", strconv.Itoa(t.GetStockCount())))
		waste := t.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("terrace.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("terrace.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		tableau := t.GetTableau()
		for pile := range domain.TerraceTableauCnt {
			cards := tableau[pile]
			b.WriteString(i18n.Tf("terrace.pileLabel", "pile", strconv.Itoa(pile)))
			if len(cards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(terracePileStr(cards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch t.GetPhase() {
		case domain.TerracePhasePlaying:
			if t.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := t.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("terrace.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(t.GetMoveCount())) + "\n")
		case domain.TerracePhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(t.GetMoveCount())) + "\n")
		case domain.TerracePhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			fnd := t.GetFoundation()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				cuiCountPileCards(fnd[:]...), domain.TerraceTotalCards)) + "\n")
		}
	})
}

// HintOutput emits the current Terrace hint.
func (p *TerraceCuiPresenter) HintOutput(t interfaces.TerraceGame) string {
	hint := t.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "reserve":
		from = i18n.T("terrace.hintFromReserve")
	case "waste":
		from = i18n.T("terrace.hintFromWaste")
	case "stock":
		from = i18n.T("terrace.hintFromStock")
	default:
		from = i18n.Tf("terrace.hintFromTableau", "pile", strconv.Itoa(hint.FromIdx))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("terrace.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	case "waste":
		to = i18n.T("terrace.hintToWaste")
	default:
		to = i18n.Tf("terrace.hintToTableau", "pile", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("terrace.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TerraceCuiPresenter) ActionLogOutput(t interfaces.TerraceGame) string {
	if t.GetPhase() == domain.TerracePhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(t.GetActionLog())
}
