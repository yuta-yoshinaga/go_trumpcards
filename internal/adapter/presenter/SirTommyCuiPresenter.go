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
			// **次に必要なランクを出す。**Web は各基礎札にバッジで出しているのに、
			// CUI は一番上の札しか出さず、4 本分の「+1」を暗算させていた (#4868)。
			// 段差もスートも無い (canPlaceOnFoundation) ので、次は必ず top+1、
			// 空なら A、13 枚で打ち止め。
			switch {
			case len(pile) >= domain.CardValueMax:
				b.WriteString(i18n.T("sirtommy.foundationComplete"))
			case len(pile) == 0:
				b.WriteString(i18n.Tf("sirtommy.foundationNext", "rank", cuiRankLabel(1)))
			default:
				b.WriteString(i18n.Tf("sirtommy.foundationNext",
					"rank", cuiRankLabel(pile[len(pile)-1].GetValue()+1)))
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
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := g.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
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
	// ファンデーションに置ける手が無い局面では、山札の札をどのウェイストに
	// 置くべきかを助言する。ここがこのゲーム唯一の戦略的判断 (#5552)。
	if hint.ToZone == "waste" {
		return i18n.Tf("sirtommy.hintPlaceWaste",
			"waste", strconv.Itoa(hint.WasteIdx)) + "\n"
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
