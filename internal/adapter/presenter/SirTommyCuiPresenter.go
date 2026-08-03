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

// SirTommyCuiPresenter renders the SirTommy Solitaire CUI view.
type SirTommyCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *SirTommyCuiPresenter) Output(g interfaces.SirTommyGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sirtommy.helpTitle"), func(b *strings.Builder) {
		// Foundations. Unlike Calculation there is no per-pile step to label:
		// every pile builds A->K one rank at a time, suit ignored, and a pile is
		// only opened by an Ace.
		foundations := g.GetFoundations()
		maxStr := strconv.Itoa(domain.CardValueMax)
		for i := range domain.SirTommyFoundationCnt {
			pile := foundations[i]
			b.WriteString(i18n.Tf("sirtommy.foundationLabel", "idx", strconv.Itoa(i)))
			if len(pile) == 0 {
				b.WriteString(i18n.T("sirtommy.foundationEmpty"))
			} else {
				top := pile[len(pile)-1]
				b.WriteString(i18n.Tf("sirtommy.foundationFilled",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(len(pile)),
					"max", maxStr))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// Wastes
		wastes := g.GetWastes()
		for i := range domain.SirTommyWasteCnt {
			pile := wastes[i]
			b.WriteString(i18n.Tf("sirtommy.wasteLabel", "idx", strconv.Itoa(i)))
			if len(pile) == 0 {
				b.WriteString(i18n.T("sirtommy.wasteEmpty"))
			} else {
				top := pile[len(pile)-1]
				b.WriteString(i18n.Tf("sirtommy.wasteFilled",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(len(pile))))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// Stock
		b.WriteString(i18n.Tf("sirtommy.stockLine",
			"count", strconv.Itoa(g.GetStockCount())))
		if top := g.GetStockTop(); top != nil {
			b.WriteString(i18n.Tf("sirtommy.stockNext", "card", cuiCardStr(top)))
		}
		b.WriteString("\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.SirTommyPhasePlaying:
			if g.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.SirTommyPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.SirTommyPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current SirTommy hint.
func (p *SirTommyCuiPresenter) HintOutput(g interfaces.SirTommyGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	switch hint.FromZone {
	case "stock":
		return i18n.Tf("sirtommy.hintStock",
			"foundation", strconv.Itoa(hint.FoundationIdx)) + "\n"
	case "waste":
		return i18n.Tf("sirtommy.hintWaste",
			"waste", strconv.Itoa(hint.WasteIdx),
			"foundation", strconv.Itoa(hint.FoundationIdx)) + "\n"
	}
	return i18n.T("cuiHintNone") + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SirTommyCuiPresenter) ActionLogOutput(g interfaces.SirTommyGame) string {
	if g.GetPhase() == domain.SirTommyPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
