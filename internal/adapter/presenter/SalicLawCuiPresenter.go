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

// saliclawPileStr returns the display string for one tableau pile.
func salicLawPileStr(pile []*domain.Card) string {
	parts := make([]string, len(pile))
	for j, card := range pile {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(card))
	}
	return strings.Join(parts, " ")
}

// SalicLawCuiPresenter renders the SalicLaw CUI view.
type SalicLawCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SalicLawCuiPresenter) Output(c interfaces.SalicLawGame, lastErr error) string {
	return buildCuiOutput(i18n.T("saliclaw.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.T("saliclaw.foundationHeader"))
		foundation := c.GetFoundation()
		for i := range domain.SalicLawFoundationCnt {
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

		// **退場したクイーンを見せる。**8 枚が場から消えている理由が盤だけでは
		// 分からないので、ゲーム名の由来ごと明示する。
		b.WriteString(i18n.T("saliclaw.queensHeader"))
		for i, q := range c.GetQueens() {
			if i != 0 {
				b.WriteString(" ")
			}
			b.WriteString(cuiCardStr(q))
		}
		b.WriteString("\n")

		b.WriteString(i18n.Tf("saliclaw.stockLine", "count", strconv.Itoa(c.GetStockCount())))
		b.WriteString("\n")

		b.WriteString("----------\n")

		tableau := c.GetTableau()
		for pile := range domain.SalicLawTableauCnt {
			cards := tableau[pile]
			b.WriteString(i18n.Tf("saliclaw.pileLabel", "pile", strconv.Itoa(pile)))
			switch {
			case len(cards) == 0:
				// まだ K が出ていない列。配りが進めば開く。
				b.WriteString(" " + i18n.T("saliclaw.emptyPile"))
			case len(cards) == 1:
				// K だけの列は、このゲームで唯一の「置ける枠」。
				b.WriteString(salicLawPileStr(cards) + " " + i18n.T("saliclaw.bareKingPile"))
			default:
				b.WriteString(salicLawPileStr(cards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch c.GetPhase() {
		case domain.SalicLawPhasePlaying:
			if c.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := c.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("saliclaw.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.SalicLawPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(c.GetMoveCount())) + "\n")
		case domain.SalicLawPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			fnd := c.GetFoundation()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				cuiCountPileCards(fnd[:]...),
				domain.SalicLawFoundationCnt*domain.SalicLawFoundationTarget)) + "\n")
		}
	})
}

// HintOutput emits the current SalicLaw hint.
func (p *SalicLawCuiPresenter) HintOutput(c interfaces.SalicLawGame) string {
	hint := c.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	// **「配れ」は移動ではない。**行き先の列を持たないので、移動の体裁
	// （A → B）に落とすと列 -1 が漏れる。専用の文言で言う。
	if hint.FromZone == "stock" {
		return i18n.T("saliclaw.hintDeal") + "\n"
	}
	from := i18n.Tf("saliclaw.hintFromTableau", "pile", strconv.Itoa(hint.FromIdx))
	to := i18n.Tf("saliclaw.hintToTableau", "pile", strconv.Itoa(hint.ToIdx))
	if hint.ToZone == "foundation" {
		to = i18n.Tf("saliclaw.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("saliclaw.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SalicLawCuiPresenter) ActionLogOutput(c interfaces.SalicLawGame) string {
	if c.GetPhase() == domain.SalicLawPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(c.GetActionLog())
}
