//go:build !js || !wasm || extra4

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BassetWebPresenter renders Basset state as the Web API response.
type BassetWebPresenter struct{}

// Output serializes the current Basset state.
func (p *BassetWebPresenter) Output(g interfaces.BassetGame, lastErr error) string {
	o := new(controller.BassetWebOutput)
	o.Phase, o.Chips = g.GetPhase(), g.GetChips()
	o.TurnsPlayed, o.TurnsTotal = g.GetTurnsPlayed(), g.GetTurnsTotal()
	o.Remaining, o.TotalPayout, o.GameEndFlag = g.GetRemainingCount(), g.GetTotalPayout(), g.GetGameEndFlag()
	r := g.GetRemainingByRank()
	o.RemainingByRank = r[:]
	if bet, rank := g.GetBet(); bet != nil {
		o.Bet = &controller.BassetWebBet{Rank: rank, Amount: bet.Amount, Stage: bet.Stage}
	}
	if turn := g.GetLastTurn(); turn != nil {
		o.BankerCard, o.PlayerCard, o.Hit = cardToOutput(turn.BankerCard), cardToOutput(turn.PlayerCard), turn.Hit
	}
	if lastErr != nil {
		o.Message = lastErr.Error()
	}
	return marshalOrError(o)
}

// ActionLogOutput serializes the Basset action log.
func (p *BassetWebPresenter) ActionLogOutput(g interfaces.BassetGame) string {
	return actionLogOutputJSON(g)
}
