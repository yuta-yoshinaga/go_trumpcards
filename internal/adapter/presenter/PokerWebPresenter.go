//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PokerWebPresenter ポーカーWebプレゼンタークラス
type PokerWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pwp *PokerWebPresenter) Output(p interfaces.PokerGame, lastErr error) string {
	resObj := pwp.buildOutput(p, lastErr)
	return marshalOrError(resObj)
}

// OutputWithOdds ゲーム状態 + ドローオッズをJSON出力
func (pwp *PokerWebPresenter) OutputWithOdds(p interfaces.PokerGame, lastErr error, odds []domain.PokerDrawOdds) string {
	resObj := pwp.buildOutput(p, lastErr)
	if odds != nil {
		resObj.Odds = make([]*controller.PokerWebOutputOdds, len(odds))
		for i, o := range odds {
			resObj.Odds[i] = &controller.PokerWebOutputOdds{
				HandRank:    o.HandRank,
				HandName:    o.HandName,
				Probability: o.Probability,
				Count:       o.Count,
				Total:       o.Total,
			}
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (pwp *PokerWebPresenter) ActionLogOutput(p interfaces.PokerGame) string {
	return actionLogOutputJSON(p)
}

// buildOutput ゲーム状態をPokerWebOutputに変換
func (pwp *PokerWebPresenter) buildOutput(p interfaces.PokerGame, lastErr error) *controller.PokerWebOutput {
	resObj := new(controller.PokerWebOutput)
	resObj.Phase = p.GetPhase()
	resObj.ExchangeRead = p.IsExchangeRead(0)
	resObj.Pot = p.GetPot()

	// **判定はドメイン。**Holdem 系と同じ形で、ベッティングフェーズのときだけ
	// 勝率とポットオッズを載せる (#4678)。
	if eq := p.GetEquity(); eq != nil {
		handOdds := make([]*controller.HoldemWebOutputHandOdds, len(eq.HandOdds))
		for i, ho := range eq.HandOdds {
			handOdds[i] = &controller.HoldemWebOutputHandOdds{
				HandRank:    ho.HandRank,
				HandName:    ho.HandName,
				Probability: ho.Probability,
			}
		}
		resObj.Equity = &controller.HoldemWebOutputEquity{WinProbability: eq.Equity, HandOdds: handOdds}
		potOdds := p.GetPotOdds()
		resObj.PotOdds = &potOdds
	}
	resObj.DealerIdx = p.GetDealerIdx()
	resObj.CurrentTurn = p.GetCurrentTurn()
	resObj.GameEndFlag = p.GetGameEndFlag()
	resObj.LastBet = p.GetLastBet()
	resObj.MinRaise = p.GetMinRaise()
	resObj.Ante = p.GetAnte()
	resObj.JokerCount = p.GetConfig().JokerCount
	resObj.BettingLimit = int(p.GetConfig().BettingLimit)
	resObj.RaiseCount = p.GetRaiseCount()
	_, resObj.MaxBetAmount = domain.CalculateBettingLimits(p.GetConfig().BettingLimit, p.GetPot(), p.GetLastBet())
	resObj.IsLowball = p.GetConfig().IsLowball

	resObj.SidePots = pwp.buildSidePotsOutput(p)
	resObj.Players = pwp.buildPlayersOutput(p)
	resObj.CpuActions = pwp.buildCpuActionsOutput(p)
	resObj.CpuExchanges = pwp.buildCpuExchangesOutput(p)
	resObj.RoundResults = pwp.buildRoundResultsOutput(p)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = pwp.buildMessage(p, lastErr)

	// メタAI情報
	if profile := p.GetHumanProfile(); profile != nil {
		resObj.MetaAI = &controller.PokerWebOutputMetaAI{
			Enabled:        true,
			GamesPlayed:    profile.GamesPlayed,
			BluffRate:      profile.BluffRate(1), // medium bracket as representative
			FoldRate:       profile.FoldRate(),
			HesitationMean: profile.HesitationMean,
		}
		d := profile.Export()
		resObj.Profile = &d
	}

	return resObj
}

// buildMessage ゲーム結果メッセージを構築
func (pwp *PokerWebPresenter) buildMessage(p interfaces.PokerGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if p.GetGameEndFlag() {
		msg, code := pwp.buildResultMessage(p)
		return msg, code, nil
	}
	return "", "", nil
}

// buildSidePotsOutput サイドポット情報を構築
func (pwp *PokerWebPresenter) buildSidePotsOutput(p interfaces.PokerGame) []*controller.PokerWebOutputSidePot {
	out := make([]*controller.PokerWebOutputSidePot, 0)
	for _, sp := range p.GetSidePots() {
		out = append(out, &controller.PokerWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (pwp *PokerWebPresenter) buildPlayersOutput(p interfaces.PokerGame) []*controller.PokerWebOutputPlayer {
	out := make([]*controller.PokerWebOutputPlayer, 0)
	isEnd := p.GetPhase() == domain.PokerPhaseEnd
	for i, player := range p.GetPlayers() {
		pObj := &controller.PokerWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Chips:         player.GetChips(),
			CurrentBet:    player.GetCurrentBet(),
			Folded:        player.GetFolded(),
			AllIn:         player.GetAllIn(),
			ExchangeCount: player.GetExchangeCount(),
			PlayStyleName: player.GetPlayStyleName(),
		}

		// 人間のカードは常に表示、CPUのカードは終了時のみ表示
		pObj.Cards = playerCardsToOutput(player, player.GetIsHuman() || (isEnd && !player.GetFolded()))

		// 終了時のハンド情報
		if isEnd && !player.GetFolded() {
			pObj.HandRank = player.GetHandRank()
			pObj.HandName = player.GetHandName()
		}

		out = append(out, pObj)
	}
	return out
}

// buildCpuActionsOutput CPU行動記録を構築
func (pwp *PokerWebPresenter) buildCpuActionsOutput(p interfaces.PokerGame) []*controller.PokerWebOutputCpuAction {
	out := make([]*controller.PokerWebOutputCpuAction, 0)
	for _, action := range p.GetCpuActions() {
		out = append(out, &controller.PokerWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}
	return out
}

// buildCpuExchangesOutput CPU交換記録を構築
func (pwp *PokerWebPresenter) buildCpuExchangesOutput(p interfaces.PokerGame) []*controller.PokerWebOutputCpuExchange {
	out := make([]*controller.PokerWebOutputCpuExchange, 0)
	for _, ex := range p.GetCpuExchanges() {
		out = append(out, &controller.PokerWebOutputCpuExchange{
			PlayerIdx:     ex.PlayerIdx,
			ExchangeCount: ex.ExchangeCount,
		})
	}
	return out
}

// buildRoundResultsOutput ラウンド結果を構築
func (pwp *PokerWebPresenter) buildRoundResultsOutput(p interfaces.PokerGame) []*controller.PokerWebOutputResult {
	out := make([]*controller.PokerWebOutputResult, 0)
	for _, r := range p.GetRoundResults() {
		out = append(out, &controller.PokerWebOutputResult{
			PlayerIdx: r.PlayerIdx,
			HandRank:  r.HandRank,
			HandName:  r.HandName,
			Kickers:   domain.FormatKickers(r.Kickers),
			WonAmount: r.WonAmount,
		})
	}
	return out
}

// buildResultMessage builds the end-of-round message and its i18n code
func (pwp *PokerWebPresenter) buildResultMessage(p interfaces.PokerGame) (string, string) {
	results := p.GetRoundResults()
	if len(results) == 0 {
		return "Game over.", "poker.result.gameOver"
	}

	players := p.GetPlayers()
	for _, r := range results {
		if players[r.PlayerIdx].GetIsHuman() {
			if r.WonAmount > 0 {
				return "You are the winner.", "poker.result.win"
			}
		}
	}

	// Human folded
	for _, pl := range players {
		if pl.GetIsHuman() && pl.GetFolded() {
			return "You folded.", "poker.result.folded"
		}
	}

	return "You lose.", "poker.result.lose"
}
