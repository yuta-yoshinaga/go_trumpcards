//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DeuceToSevenWebPresenter renders 2-7 Triple Draw game state as a JSON payload.
type DeuceToSevenWebPresenter struct{}

// Output marshals the current game state to JSON.
func (dwp *DeuceToSevenWebPresenter) Output(g interfaces.DeuceToSevenGame, lastErr error) string {
	return marshalOrError(dwp.buildOutput(g, lastErr))
}

// ActionLogOutput marshals the action log to JSON.
func (dwp *DeuceToSevenWebPresenter) ActionLogOutput(g interfaces.DeuceToSevenGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput ヒントを出力する。Web ではヒントはクライアント側 (useGameHint) で
// 算出するため、通常の状態出力を返す。DeuceToSevenPresenter インタフェースを満たすための実装。
func (dwp *DeuceToSevenWebPresenter) HintOutput(g interfaces.DeuceToSevenGame) string {
	return dwp.Output(g, nil)
}

func (dwp *DeuceToSevenWebPresenter) buildOutput(g interfaces.DeuceToSevenGame, lastErr error) *controller.DeuceToSevenWebOutput {
	out := new(controller.DeuceToSevenWebOutput)
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

	out.SidePots = dwp.buildSidePots(g)
	out.Players = dwp.buildPlayers(g)
	out.CpuActions = dwp.buildCpuActions(g)
	out.CpuExchanges = dwp.buildCpuExchanges(g)
	out.RoundResults = dwp.buildRoundResults(g)
	out.Message, out.MessageCode, out.MessageParams = dwp.buildMessage(g, lastErr)

	if profile := g.GetHumanProfile(); profile != nil {
		out.MetaAI = &controller.DeuceToSevenWebOutputMetaAI{
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

func (dwp *DeuceToSevenWebPresenter) buildMessage(g interfaces.DeuceToSevenGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		msg, code := dwp.buildResultMessage(g)
		return msg, code, nil
	}
	return "", "", nil
}

func (dwp *DeuceToSevenWebPresenter) buildResultMessage(g interfaces.DeuceToSevenGame) (string, string) {
	results := g.GetRoundResults()
	if len(results) == 0 {
		return "Game over.", "deucetoseven.result.gameOver"
	}
	players := g.GetPlayers()
	for _, r := range results {
		if players[r.PlayerIdx].GetIsHuman() && r.WonAmount > 0 {
			return "You are the winner.", "deucetoseven.result.win"
		}
	}
	for _, pl := range players {
		if pl.GetIsHuman() && pl.GetFolded() {
			return "You folded.", "deucetoseven.result.folded"
		}
	}
	return "You lose.", "deucetoseven.result.lose"
}

func (dwp *DeuceToSevenWebPresenter) buildSidePots(g interfaces.DeuceToSevenGame) []*controller.DeuceToSevenWebOutputSidePot {
	out := make([]*controller.DeuceToSevenWebOutputSidePot, 0)
	for _, sp := range g.GetSidePots() {
		out = append(out, &controller.DeuceToSevenWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

func (dwp *DeuceToSevenWebPresenter) buildPlayers(g interfaces.DeuceToSevenGame) []*controller.DeuceToSevenWebOutputPlayer {
	out := make([]*controller.DeuceToSevenWebOutputPlayer, 0)
	isEnd := g.GetPhase() == domain.DeuceToSevenPhaseEnd
	for i, pl := range g.GetPlayers() {
		obj := &controller.DeuceToSevenWebOutputPlayer{
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
			obj.HandRank = pl.GetHandRank()
			obj.HandName = pl.GetHandName()
		}
		out = append(out, obj)
	}
	return out
}

func (dwp *DeuceToSevenWebPresenter) buildCpuActions(g interfaces.DeuceToSevenGame) []*controller.DeuceToSevenWebOutputCpuAction {
	out := make([]*controller.DeuceToSevenWebOutputCpuAction, 0)
	for _, a := range g.GetCpuActions() {
		out = append(out, &controller.DeuceToSevenWebOutputCpuAction{
			PlayerIdx:  a.PlayerIdx,
			Action:     a.Action,
			Amount:     a.Amount,
			DrawIndex:  a.DrawIndex,
			RoundLabel: a.RoundLabel,
		})
	}
	return out
}

func (dwp *DeuceToSevenWebPresenter) buildCpuExchanges(g interfaces.DeuceToSevenGame) []*controller.DeuceToSevenWebOutputCpuExchange {
	out := make([]*controller.DeuceToSevenWebOutputCpuExchange, 0)
	for _, e := range g.GetCpuExchanges() {
		out = append(out, &controller.DeuceToSevenWebOutputCpuExchange{
			PlayerIdx:     e.PlayerIdx,
			DrawIndex:     e.DrawIndex,
			ExchangeCount: e.ExchangeCount,
		})
	}
	return out
}

func (dwp *DeuceToSevenWebPresenter) buildRoundResults(g interfaces.DeuceToSevenGame) []*controller.DeuceToSevenWebOutputResult {
	out := make([]*controller.DeuceToSevenWebOutputResult, 0)
	for _, r := range g.GetRoundResults() {
		out = append(out, &controller.DeuceToSevenWebOutputResult{
			PlayerIdx: r.PlayerIdx,
			HandRank:  r.HandRank,
			HandName:  r.HandName,
			WonAmount: r.WonAmount,
		})
	}
	return out
}
