//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// slyfoxPileStr returns the display string for one tableau pile.
//
// **動かせるのは一番上の 1 枚だけ。**`m t <山>` は山番号しか取らないのに、
// 以前は全カードに [0][1][2]… と添字を振っていて、埋まった札まで番号で
// 指定できるように見せていた (#5739)。番号を振るのをやめ、一番上だけ
// 印を付ける。
func slyFoxPileStr(pile []*domain.Card) string {
	parts := make([]string, len(pile))
	for j, card := range pile {
		if j == len(pile)-1 {
			parts[j] = " " + i18n.Tf("slyfox.pileTop", "card", cuiCardStr(card))
			continue
		}
		parts[j] = " " + i18n.Tf("slyfox.pileBuried", "card", cuiCardStr(card))
	}
	return strings.Join(parts, " ")
}

// slyFoxDirMark returns the arrow that tells an ascending foundation from a
// descending one.
func slyFoxDirMark(ascending bool) string {
	if ascending {
		return "\u2191"
	}
	return "\u2193"
}

// SlyFoxCuiPresenter renders the SlyFox CUI view.
type SlyFoxCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SlyFoxCuiPresenter) Output(c interfaces.SlyFoxGame, lastErr error) string {
	return buildCuiOutput(i18n.T("slyfox.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.T("slyfox.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.SlyFoxFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			// 8 つのうちどれが A→K でどれが K→A かは、表示しないと盤面から読めない。
			b.WriteString(slyFoxDirMark(c.IsAscendingFoundation(i)))
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		b.WriteString(i18n.Tf("slyfox.stockLine", "count", strconv.Itoa(c.GetStockCount())))
		b.WriteString("\n")
		// **あと何枚でリザーブが開くかを言う。**閉じている理由は盤からは
		// 読めないので、書かないと「なぜ送れないのか」が分からない。
		if c.ReserveIsLocked() {
			b.WriteString(i18n.Tf("slyfox.reserveLocked",
				"dealt", strconv.Itoa(c.DealtThisCycle()),
				"cycle", strconv.Itoa(domain.SlyFoxDealCycle),
				"left", strconv.Itoa(domain.SlyFoxDealCycle-c.DealtThisCycle())))
		} else {
			b.WriteString(i18n.T("slyfox.reserveOpen"))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		b.WriteString(i18n.T("slyfox.pileTopNote") + "\n")

		tableau := c.GetTableau()
		for pile := range domain.SlyFoxTableauCnt {
			cards := tableau[pile]
			b.WriteString(i18n.Tf("slyfox.pileLabel", "pile", strconv.Itoa(pile)))
			if len(cards) == 0 {
				// 空き枠は補充されない。次に配るときの置き場所になるだけ。
				b.WriteString(" " + i18n.T("slyfox.emptyPile"))
			} else {
				b.WriteString(slyFoxPileStr(cards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.SlyFoxPhasePlaying:
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := c.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("slyfox.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.SlyFoxPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.SlyFoxPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current SlyFox hint.
func (p *SlyFoxCuiPresenter) HintOutput(c interfaces.SlyFoxGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	// 捨て札は無いので、移動元は山札かリザーブのどちらかしかない。
	from := i18n.Tf("slyfox.hintFromTableau", "pile", strconv.Itoa(hint.FromIdx))
	if hint.FromZone == "stock" {
		from = i18n.T("slyfox.hintFromStock")
	}
	to := i18n.Tf("slyfox.hintToTableau", "pile", strconv.Itoa(hint.ToIdx))
	if hint.ToZone == "foundation" {
		to = i18n.Tf("slyfox.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("slyfox.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SlyFoxCuiPresenter) ActionLogOutput(c interfaces.SlyFoxGame) string {
	if c.GetPhase() == domain.SlyFoxPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
