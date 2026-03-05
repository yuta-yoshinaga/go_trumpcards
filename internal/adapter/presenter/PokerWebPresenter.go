package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PokerWebPresenter ポーカーWebプレゼンタークラス
type PokerWebPresenter struct{}

// NewPokerWebPresenter コンストラクタ
func NewPokerWebPresenter() *PokerWebPresenter {
	return &PokerWebPresenter{}
}

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

// buildOutput ゲーム状態をPokerWebOutputに変換
func (pwp *PokerWebPresenter) buildOutput(p interfaces.PokerGame, lastErr error) *controller.PokerWebOutput {
	resObj := new(controller.PokerWebOutput)
	resObj.Phase = p.GetPhase()
	resObj.Pot = p.GetPot()
	resObj.DealerIdx = p.GetDealerIdx()
	resObj.CurrentTurn = p.GetCurrentTurn()
	resObj.GameEndFlag = p.GetGameEndFlag()
	resObj.LastBet = p.GetLastBet()
	resObj.MinRaise = p.GetMinRaise()
	resObj.Ante = p.GetAnte()
	resObj.JokerCount = p.GetConfig().JokerCount

	// サイドポット
	resObj.SidePots = make([]*controller.PokerWebOutputSidePot, 0)
	for _, sp := range p.GetSidePots() {
		resObj.SidePots = append(resObj.SidePots, &controller.PokerWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}

	// プレイヤー情報
	resObj.Players = make([]*controller.PokerWebOutputPlayer, 0)
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
			Cards:         make([]*controller.WebOutputCard, 0),
		}

		// 人間のカードは常に表示、CPUのカードは終了時のみ表示
		if player.GetIsHuman() || (isEnd && !player.GetFolded()) {
			for j := 0; j < player.GetCardsSize(); j++ {
				pObj.Cards = append(pObj.Cards, cardToOutput(player.GetCard(j)))
			}
		}

		// 終了時のハンド情報
		if isEnd && !player.GetFolded() {
			pObj.HandRank = player.GetHandRank()
			pObj.HandName = player.GetHandName()
		}

		resObj.Players = append(resObj.Players, pObj)
	}

	// CPU行動記録
	resObj.CpuActions = make([]*controller.PokerWebOutputCpuAction, 0)
	for _, action := range p.GetCpuActions() {
		resObj.CpuActions = append(resObj.CpuActions, &controller.PokerWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}

	// CPU交換記録
	resObj.CpuExchanges = make([]*controller.PokerWebOutputCpuExchange, 0)
	for _, ex := range p.GetCpuExchanges() {
		resObj.CpuExchanges = append(resObj.CpuExchanges, &controller.PokerWebOutputCpuExchange{
			PlayerIdx:     ex.PlayerIdx,
			ExchangeCount: ex.ExchangeCount,
		})
	}

	// ラウンド結果
	resObj.RoundResults = make([]*controller.PokerWebOutputResult, 0)
	for _, r := range p.GetRoundResults() {
		resObj.RoundResults = append(resObj.RoundResults, &controller.PokerWebOutputResult{
			PlayerIdx: r.PlayerIdx,
			HandRank:  r.HandRank,
			HandName:  r.HandName,
			WonAmount: r.WonAmount,
		})
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if p.GetGameEndFlag() {
		msg, code := pwp.buildResultMessage(p)
		resObj.Message = msg
		resObj.MessageCode = code
	}

	return resObj
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
