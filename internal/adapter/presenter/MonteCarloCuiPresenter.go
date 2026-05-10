package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// MonteCarloCuiPresenter renders the Monte Carlo Solitaire CUI view.
type MonteCarloCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (pr *MonteCarloCuiPresenter) Output(g interfaces.MonteCarloGame, lastErr error) string {
	return buildCuiOutput(i18n.T("montecarlo.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("montecarlo.header",
			"stock", strconv.Itoa(g.GetStockCount()),
			"removed", strconv.Itoa(g.GetRemovedCount()),
			"deals", strconv.Itoa(g.GetDealCount())) + "\n")
		b.WriteString("----------\n")

		board := g.GetBoard()
		for r := range domain.MonteCarloGridSize {
			for c := range domain.MonteCarloGridSize {
				if c > 0 {
					b.WriteString(" | ")
				}
				rs := strconv.Itoa(r)
				cs := strconv.Itoa(c)
				if card := board[r][c]; card == nil {
					b.WriteString(i18n.Tf("montecarlo.cellEmpty", "r", rs, "c", cs))
				} else {
					b.WriteString(i18n.Tf("montecarlo.cellCard",
						"r", rs, "c", cs, "card", cuiCardStr(card)))
				}
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.MonteCarloPhasePlaying:
			if g.IsStalemate() {
				b.WriteString(color.Red(i18n.T("montecarlo.stalemate")) + "\n")
			}
		case domain.MonteCarloPhaseGameClear:
			b.WriteString(color.Green(i18n.Tf("montecarlo.gameClear",
				"dealCount", strconv.Itoa(g.GetDealCount()))) + "\n")
		case domain.MonteCarloPhaseGameOver:
			b.WriteString(color.Red(i18n.T("montecarlo.gameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Monte Carlo hint.
func (pr *MonteCarloCuiPresenter) HintOutput(g interfaces.MonteCarloGame) string {
	hint := g.Hint()
	if hint == nil {
		return i18n.T("montecarlo.noHint") + "\n"
	}
	if hint.Action == domain.MonteCarloHintActionDeal {
		return i18n.T("montecarlo.hintLineDeal") + "\n"
	}
	return i18n.Tf("montecarlo.hintLineRemove",
		"r1", strconv.Itoa(hint.FromR),
		"c1", strconv.Itoa(hint.FromC),
		"r2", strconv.Itoa(hint.ToR),
		"c2", strconv.Itoa(hint.ToC)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *MonteCarloCuiPresenter) ActionLogOutput(g interfaces.MonteCarloGame) string {
	if g.GetPhase() == domain.MonteCarloPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
