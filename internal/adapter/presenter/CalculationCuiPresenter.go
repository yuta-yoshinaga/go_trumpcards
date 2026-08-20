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

// CalculationCuiPresenter renders the Calculation Solitaire CUI view.
type CalculationCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *CalculationCuiPresenter) Output(g interfaces.CalculationGame, lastErr error) string {
	return buildCuiOutput(i18n.T("calculation.helpTitle"), func(b *strings.Builder) {
		// Foundations (one per step: +1, +2, +3, +4)
		foundations := g.GetFoundations()
		stepLabels := []string{"+1", "+2", "+3", "+4"}
		maxStr := strconv.Itoa(domain.CardValueMax)
		for i := range domain.CalculationFoundationCnt {
			pile := foundations[i]
			b.WriteString(i18n.Tf("calculation.foundationLabel",
				"idx", strconv.Itoa(i),
				"step", stepLabels[i]))
			if len(pile) == 0 {
				b.WriteString(i18n.T("calculation.foundationEmpty"))
			} else {
				top := pile[len(pile)-1]
				b.WriteString(i18n.Tf("calculation.foundationFilled",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(len(pile)),
					"max", maxStr))
			}
			// **各列が +1/+2/+3/+4 ずつ 13 を法として進む (#4794)。**Web は
			// 次に置けるランクをバッジで常時出しているのに、CUI は一番上の札
			// しか出さず、毎手この暗算を強いていた。
			if next := g.GetNextFoundationRank(i); next > 0 {
				b.WriteString(i18n.Tf("calculation.nextRank",
					"rank", strconv.Itoa(next)))
			}
			// 1手先だけでは +3 の列で「次の次のまた次」を辿れない。Web は6手先まで
			// バッジで出している (#5551)。空の組札には出さない。
			if upcoming := g.GetUpcomingFoundationRanks(i, domain.CalculationMaxLookAhead); len(upcoming) > 0 {
				parts := make([]string, len(upcoming))
				for j, r := range upcoming {
					parts[j] = strconv.Itoa(r)
				}
				b.WriteString(i18n.Tf("calculation.upcomingRanks",
					"ranks", strings.Join(parts, " → ")))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// Wastes
		wastes := g.GetWastes()
		for i := range domain.CalculationWasteCnt {
			pile := wastes[i]
			b.WriteString(i18n.Tf("calculation.wasteLabel", "idx", strconv.Itoa(i)))
			if len(pile) == 0 {
				b.WriteString(i18n.T("calculation.wasteEmpty"))
			} else {
				top := pile[len(pile)-1]
				b.WriteString(i18n.Tf("calculation.wasteFilled",
					"card", cuiCardStr(top),
					"count", strconv.Itoa(len(pile))))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		// Stock
		b.WriteString(i18n.Tf("calculation.stockLine",
			"count", strconv.Itoa(g.GetStockCount())))
		if top := g.GetStockTop(); top != nil {
			b.WriteString(i18n.Tf("calculation.stockNext", "card", cuiCardStr(top)))
		}
		b.WriteString("\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.CalculationPhasePlaying:
			if g.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := g.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.CalculationPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.CalculationPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Calculation hint.
func (p *CalculationCuiPresenter) HintOutput(g interfaces.CalculationGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	switch hint.FromZone {
	case "stock":
		return i18n.Tf("calculation.hintStock",
			"foundation", strconv.Itoa(hint.FoundationIdx)) + "\n"
	case "waste":
		return i18n.Tf("calculation.hintWaste",
			"waste", strconv.Itoa(hint.WasteIdx),
			"foundation", strconv.Itoa(hint.FoundationIdx)) + "\n"
	}
	return i18n.T("cuiHintNone") + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CalculationCuiPresenter) ActionLogOutput(g interfaces.CalculationGame) string {
	if g.GetPhase() == domain.CalculationPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
