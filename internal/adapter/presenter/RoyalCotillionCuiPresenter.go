//go:build !js || !wasm || classic

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

// royalcotillionPileStr returns the display string for one tableau pile.
func royalcotillionPileStr(pile []*domain.Card) string {
	parts := make([]string, len(pile))
	for j, card := range pile {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(card))
	}
	return strings.Join(parts, " ")
}

// royalCotillionSeriesMark tells an Ace-start foundation from a deuce-start one.
func royalCotillionSeriesMark(odd bool) string {
	if odd {
		return "A:"
	}
	return "2:"
}

// RoyalCotillionCuiPresenter renders the RoyalCotillion CUI view.
type RoyalCotillionCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *RoyalCotillionCuiPresenter) Output(c interfaces.RoyalCotillionGame, lastErr error) string {
	return buildCuiOutput(i18n.T("royalcotillion.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.T("royalcotillion.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.RoyalCotillionFoundationCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			// A 始まりか 2 始まりかは、表示しないと盤面から読めない。
			b.WriteString(royalCotillionSeriesMark(c.IsOddFoundation(i)))
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		b.WriteString(i18n.Tf("royalcotillion.stockLine", "count", strconv.Itoa(c.GetStockCount())))
		waste := c.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("royalcotillion.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("royalcotillion.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// タブローは 1 枠 1 枚。空き枠は山札か捨て札から補充される。
		tableau := c.GetTableau()
		for slot := range domain.RoyalCotillionTableauCnt {
			b.WriteString(i18n.Tf("royalcotillion.slotLabel", "slot", strconv.Itoa(slot)))
			if card := tableau[slot]; card == nil {
				b.WriteString(" " + i18n.T("royalcotillion.emptySlot"))
			} else {
				b.WriteString(" " + cuiCardStr(card))
			}
			if slot%4 == 3 {
				b.WriteString("\n")
			}
		}

		b.WriteString("----------\n")

		// リザーブは一番上だけが使え、空いた山は二度と埋まらない。
		reserve := c.GetReserve()
		for pile := range domain.RoyalCotillionReserveCnt {
			b.WriteString(i18n.Tf("royalcotillion.reserveLabel", "pile", strconv.Itoa(pile)))
			if cards := reserve[pile]; len(cards) == 0 {
				b.WriteString(" " + i18n.T("royalcotillion.emptyReserve"))
			} else {
				b.WriteString(royalcotillionPileStr(cards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.RoyalCotillionPhasePlaying:
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := c.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("royalcotillion.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.RoyalCotillionPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.RoyalCotillionPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current RoyalCotillion hint.
func (p *RoyalCotillionCuiPresenter) HintOutput(c interfaces.RoyalCotillionGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "waste":
		from = i18n.T("royalcotillion.hintFromWaste")
	case "stock":
		from = i18n.T("royalcotillion.hintFromStock")
	case "reserve":
		from = i18n.Tf("royalcotillion.hintFromReserve", "pile", strconv.Itoa(hint.FromIdx))
	default:
		from = i18n.Tf("royalcotillion.hintFromTableau", "pile", strconv.Itoa(hint.FromIdx))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("royalcotillion.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	case "waste":
		to = i18n.T("royalcotillion.hintToWaste")
	default:
		to = i18n.Tf("royalcotillion.hintToTableau", "pile", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("royalcotillion.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *RoyalCotillionCuiPresenter) ActionLogOutput(c interfaces.RoyalCotillionGame) string {
	if c.GetPhase() == domain.RoyalCotillionPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
