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

// AuldLangSyneCuiPresenter renders the Auld Lang Syne CUI view.
type AuldLangSyneCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *AuldLangSyneCuiPresenter) Output(g interfaces.AuldLangSyneGame, lastErr error) string {
	return buildCuiOutput(i18n.T("auldlangsyne.helpTitle"), func(b *strings.Builder) {
		// Foundations. Every pile builds A->K one rank at a time with suit
		// ignored, so the only thing worth showing per pile is its top card and
		// the rank it wants next -- the same "don't make the player add one in
		// their head" fix Sir Tommy got in #4868.
		foundations := g.GetFoundations()
		maxStr := strconv.Itoa(domain.CardValueMax)
		for i := range domain.AuldLangSyneFoundationCnt {
			pile := foundations[i]
			b.WriteString(i18n.Tf("auldlangsyne.foundationLabel", "idx", strconv.Itoa(i)))
			if len(pile) == 0 {
				b.WriteString(i18n.T("auldlangsyne.foundationEmpty"))
			} else {
				top := pile[len(pile)-1]
				b.WriteString(i18n.Tf("auldlangsyne.foundationFilled",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(len(pile)),
					"max", maxStr))
			}
			switch {
			case len(pile) >= domain.CardValueMax:
				b.WriteString(i18n.T("auldlangsyne.foundationComplete"))
			case len(pile) == 0:
				b.WriteString(i18n.Tf("auldlangsyne.foundationNext", "rank", cuiRankLabel(1)))
			default:
				b.WriteString(i18n.Tf("auldlangsyne.foundationNext",
					"rank", cuiRankLabel(pile[len(pile)-1].GetValue()+1)))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// Wastes. Only the top card is movable, so the buried count is what the
		// player needs in order to judge how badly a pile is blocked.
		wastes := g.GetWastes()
		for i := range domain.AuldLangSyneWasteCnt {
			pile := wastes[i]
			b.WriteString(i18n.Tf("auldlangsyne.wasteLabel", "idx", strconv.Itoa(i)))
			if len(pile) == 0 {
				b.WriteString(i18n.T("auldlangsyne.wasteEmpty"))
			} else {
				top := pile[len(pile)-1]
				b.WriteString(i18n.Tf("auldlangsyne.wasteFilled",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(len(pile))))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// Stock. There is no "next card" line here as there is in Sir Tommy: the
		// deal is forced onto all four wastes at once, so the player never sees
		// the next card before committing to it. What matters is how many deals
		// are left, which is the count divided by the four wastes.
		b.WriteString(i18n.Tf("auldlangsyne.stockLine",
			"count", strconv.Itoa(g.GetStockCount()),
			"deals", strconv.Itoa(g.GetStockCount()/domain.AuldLangSyneWasteCnt)))
		b.WriteString("\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.AuldLangSynePhasePlaying:
			if g.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.AuldLangSynePhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.AuldLangSynePhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Auld Lang Syne hint.
func (p *AuldLangSyneCuiPresenter) HintOutput(g interfaces.AuldLangSyneGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	return i18n.Tf("auldlangsyne.hintWaste",
		"waste", strconv.Itoa(hint.WasteIdx),
		"foundation", strconv.Itoa(hint.FoundationIdx)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *AuldLangSyneCuiPresenter) ActionLogOutput(g interfaces.AuldLangSyneGame) string {
	if g.GetPhase() == domain.AuldLangSynePhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
