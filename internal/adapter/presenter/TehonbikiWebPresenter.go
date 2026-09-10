//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TehonbikiWebPresenter serializes the number game state.
type TehonbikiWebPresenter struct{}

func (*TehonbikiWebPresenter) Output(c interfaces.TehonbikiGame, e error) string {
	o := &controller.TehonbikiWebOutput{Phase: int(c.GetPhase()), Numbers: c.GetNumbers(), BetType: c.GetBetType(), Bet: c.GetBet(), Result: int(c.GetResult()), Payout: c.GetPayout(), Chips: c.GetChips(), RoundNumber: c.GetRoundNumber(), RemainingCards: c.GetRemainingCards(), GameEndFlag: c.GetGameEndFlag()}
	if c.GetPhase() == domain.TehonbikiPhaseResult || c.GetGameEndFlag() {
		n := c.GetParentCard()
		o.ParentCard = &n
	}
	if p, ok := domain.TehonbikiPayouts[c.GetBetType()]; ok {
		o.PayoutNum = p.MultiplierNum
		o.PayoutDen = p.MultiplierDen
	}
	cfg := c.GetConfig()
	o.Config = &controller.TehonbikiWebOutCfg{InitialChips: cfg.InitialChips, DefaultBet: cfg.DefaultBet}
	if e != nil {
		o.Message = e.Error()
	}
	return marshalOrError(o)
}
func (*TehonbikiWebPresenter) ActionLogOutput(c interfaces.TehonbikiGame) string {
	return actionLogOutputJSON(c)
}
func (*TehonbikiWebPresenter) HintOutput(c interfaces.TehonbikiGame) string {
	if h := c.GetHint(); h != nil {
		return marshalOrError(map[string]any{"reason": h.Reason})
	}
	return marshalOrError(map[string]any{"hint": nil})
}
