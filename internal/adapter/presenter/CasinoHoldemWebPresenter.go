//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CasinoHoldemWebPresenter カジノホールデムWebプレゼンタークラス
type CasinoHoldemWebPresenter struct{}

// Output ゲーム状態を出力
func (cp *CasinoHoldemWebPresenter) Output(g interfaces.CasinoHoldemGame, lastErr error) string {
	resObj := new(controller.CasinoHoldemWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(g.GetPlayerHand())
	if g.GetPhase() == domain.CasinoHoldemPhaseEnd && g.GetCallBet() > 0 {
		// ショーダウン到達時のみディーラーホールを公開する。フォールド時は公開しない。
		resObj.DealerHand = cardsToOutputOrEmpty(g.GetDealerHand())
	} else {
		resObj.DealerHand = casinoHoldemMaskDealerHand(g.GetDealerHand())
	}
	resObj.Community = cardsToOutputOrEmpty(g.GetCommunity())
	resObj.Phase = g.GetPhase()
	resObj.Chips = g.GetChips()
	resObj.AnteBet = g.GetAnteBet()
	resObj.BonusBet = g.GetBonusBet()
	resObj.CallBet = g.GetCallBet()
	resObj.Result = int(g.GetResult())
	resObj.DealerQualify = g.GetDealerQualify()
	resObj.AntePayout = g.GetAntePayout()
	resObj.CallPayout = g.GetCallPayout()
	resObj.BonusPayout = g.GetBonusPayout()
	resObj.TotalPayout = g.GetTotalPayout()
	resObj.PlayerHandRank = g.GetPlayerHandRank()
	resObj.DealerHandRank = g.GetDealerHandRank()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() {
		switch g.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "casinoholdem.result.playerWins"
		case domain.GameResultLose:
			if g.GetCallBet() == 0 {
				resObj.Message = "Player folded."
				resObj.MessageCode = "casinoholdem.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "casinoholdem.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "casinoholdem.result.push"
		default:
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (cp *CasinoHoldemWebPresenter) ActionLogOutput(g interfaces.CasinoHoldemGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput はヒントを返す。Web ではクライアント側でヒントを算出するため、
// 状態出力にフォールバックする (CUI プレゼンターのみが専用ヒントを返す)。
func (cp *CasinoHoldemWebPresenter) HintOutput(g interfaces.CasinoHoldemGame) string {
	return cp.Output(g, nil)
}

// casinoHoldemMaskDealerHand returns the dealer hand with all cards masked
// while the showdown has not yet happened.
func casinoHoldemMaskDealerHand(cards []*domain.Card) []*controller.WebOutputCard {
	if len(cards) == 0 {
		return make([]*controller.WebOutputCard, 0)
	}
	result := make([]*controller.WebOutputCard, len(cards))
	for i := range cards {
		result[i] = &controller.WebOutputCard{Design: "", Value: 0}
	}
	return result
}
