//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BadugiWebPresenter renders Badugi game state as a JSON payload.
type BadugiWebPresenter struct{}

// Output marshals the current game state to JSON.
func (bwp *BadugiWebPresenter) Output(g interfaces.BadugiGame, lastErr error) string {
	return marshalOrError(bwp.buildOutput(g, lastErr))
}

// ActionLogOutput marshals the action log to JSON.
// HintOutput returns the current state as JSON. The Web GUI computes its own
// draw hint client-side, so this mirrors Output to satisfy the BadugiPresenter
// interface shared with the CUI.
func (bwp *BadugiWebPresenter) HintOutput(g interfaces.BadugiGame) string {
	return bwp.Output(g, nil)
}

func (bwp *BadugiWebPresenter) ActionLogOutput(g interfaces.BadugiGame) string {
	return actionLogOutputJSON(g)
}

func (bwp *BadugiWebPresenter) buildOutput(g interfaces.BadugiGame, lastErr error) *controller.BadugiWebOutput {
	out := new(controller.BadugiWebOutput)
	out.Phase = g.GetPhase()
	out.DrawIndex = g.GetDrawIndex()
	out.Pot = g.GetPot()
	out.DealerIdx = g.GetDealerIdx()
	out.CurrentTurn = g.GetCurrentTurn()
	out.GameEndFlag = g.GetGameEndFlag()
	out.LastBet = g.GetLastBet()
	out.MinRaise = g.GetMinRaise()
	out.Ante = g.GetAnte()
	out.BettingLimit = int(g.GetConfig().BettingLimit)
	out.RaiseCount = g.GetRaiseCount()
	_, out.MaxBetAmount = domain.CalculateBettingLimits(g.GetConfig().BettingLimit, g.GetPot(), g.GetLastBet())

	out.SidePots = bwp.buildSidePots(g)
	out.Players = bwp.buildPlayers(g)
	out.CpuActions = bwp.buildCpuActions(g)
	out.CpuExchanges = bwp.buildCpuExchanges(g)
	out.RoundResults = bwp.buildRoundResults(g)
	out.Message, out.MessageCode, out.MessageParams = bwp.buildMessage(g, lastErr)

	if profile := g.GetHumanProfile(); profile != nil {
		out.MetaAI = &controller.BadugiWebOutputMetaAI{
			Enabled:        true,
			GamesPlayed:    profile.GamesPlayed,
			BluffRate:      profile.BluffRate(1),
			FoldRate:       profile.FoldRate(),
			HesitationMean: profile.HesitationMean,
		}
		d := profile.Export()
		out.Profile = &d
	}

	return out
}

func (bwp *BadugiWebPresenter) buildMessage(g interfaces.BadugiGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		msg, code := bwp.buildResultMessage(g)
		return msg, code, nil
	}
	return "", "", nil
}

func (bwp *BadugiWebPresenter) buildResultMessage(g interfaces.BadugiGame) (string, string) {
	results := g.GetRoundResults()
	if len(results) == 0 {
		return "", "badugi.result.gameOver"
	}
	players := g.GetPlayers()
	for _, r := range results {
		if players[r.PlayerIdx].GetIsHuman() && r.WonAmount > 0 {
			return "", "badugi.result.win"
		}
	}
	for _, pl := range players {
		if pl.GetIsHuman() && pl.GetFolded() {
			return "", "badugi.result.folded"
		}
	}
	return "", "badugi.result.lose"
}

func (bwp *BadugiWebPresenter) buildSidePots(g interfaces.BadugiGame) []*controller.BadugiWebOutputSidePot {
	out := make([]*controller.BadugiWebOutputSidePot, 0)
	for _, sp := range g.GetSidePots() {
		out = append(out, &controller.BadugiWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

func (bwp *BadugiWebPresenter) buildPlayers(g interfaces.BadugiGame) []*controller.BadugiWebOutputPlayer {
	out := make([]*controller.BadugiWebOutputPlayer, 0)
	isEnd := g.GetPhase() == domain.BadugiPhaseEnd
	for i, pl := range g.GetPlayers() {
		obj := &controller.BadugiWebOutputPlayer{
			ID:            i,
			IsHuman:       pl.GetIsHuman(),
			Chips:         pl.GetChips(),
			CurrentBet:    pl.GetCurrentBet(),
			Folded:        pl.GetFolded(),
			AllIn:         pl.GetAllIn(),
			DrawCount:     pl.GetDrawCount(),
			TotalDraws:    pl.GetTotalDrawCount(),
			PlayStyleName: pl.GetPlayStyleName(),
		}
		// Always show the human's hand; CPUs only reveal on showdown.
		obj.Cards = playerCardsToOutput(pl, pl.GetIsHuman() || (isEnd && !pl.GetFolded()))
		if isEnd && !pl.GetFolded() {
			obj.HandSize = pl.GetHandRank()
			obj.HandName = pl.GetHandName()
			// Include the best-subset view so the UI can highlight the selected cards.
			best := pl.GetBestHand()
			if len(best.Cards) > 0 {
				obj.BestCards = cardsToOutput(best.Cards)
			}
		}
		out = append(out, obj)
	}
	return out
}

func (bwp *BadugiWebPresenter) buildCpuActions(g interfaces.BadugiGame) []*controller.BadugiWebOutputCpuAction {
	out := make([]*controller.BadugiWebOutputCpuAction, 0)
	for _, a := range g.GetCpuActions() {
		out = append(out, &controller.BadugiWebOutputCpuAction{
			PlayerIdx:  a.PlayerIdx,
			Action:     a.Action,
			Amount:     a.Amount,
			DrawIndex:  a.DrawIndex,
			RoundLabel: a.RoundLabel,
		})
	}
	return out
}

func (bwp *BadugiWebPresenter) buildCpuExchanges(g interfaces.BadugiGame) []*controller.BadugiWebOutputCpuExchange {
	out := make([]*controller.BadugiWebOutputCpuExchange, 0)
	for _, e := range g.GetCpuExchanges() {
		out = append(out, &controller.BadugiWebOutputCpuExchange{
			PlayerIdx:     e.PlayerIdx,
			DrawIndex:     e.DrawIndex,
			ExchangeCount: e.ExchangeCount,
		})
	}
	return out
}

func (bwp *BadugiWebPresenter) buildRoundResults(g interfaces.BadugiGame) []*controller.BadugiWebOutputResult {
	out := make([]*controller.BadugiWebOutputResult, 0)
	for _, r := range g.GetRoundResults() {
		out = append(out, &controller.BadugiWebOutputResult{
			PlayerIdx: r.PlayerIdx,
			HandSize:  r.HandSize,
			HandName:  r.HandName,
			WonAmount: r.WonAmount,
		})
	}
	return out
}
