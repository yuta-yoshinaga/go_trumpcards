//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// UltimateTexasHoldemWebPresenter アルティメット・テキサスホールデムWebプレゼンタークラス
type UltimateTexasHoldemWebPresenter struct{}

// Output ゲーム状態を出力
func (up *UltimateTexasHoldemWebPresenter) Output(g interfaces.UltimateTexasHoldemGame, lastErr error) string {
	resObj := new(controller.UltimateTexasHoldemWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(g.GetPlayerHand())
	if g.GetPhase() == domain.UltimateTexasHoldemPhaseEnd {
		resObj.DealerHand = cardsToOutputOrEmpty(g.GetDealerHand())
	} else {
		resObj.DealerHand = ultimateTexasHoldemMaskDealerHand(g.GetDealerHand())
	}
	resObj.Community = cardsToOutputOrEmpty(g.GetCommunity())
	resObj.Phase = g.GetPhase()
	resObj.Chips = g.GetChips()
	resObj.AnteBet = g.GetAnteBet()
	resObj.BlindBet = g.GetBlindBet()
	resObj.TripsBet = g.GetTripsBet()
	resObj.PlayBet = g.GetPlayBet()
	resObj.Folded = g.GetFolded()
	resObj.Result = int(g.GetResult())
	resObj.DealerQualified = g.GetDealerQualified()
	resObj.AntePayout = g.GetAntePayout()
	resObj.BlindPayout = g.GetBlindPayout()
	resObj.PlayPayout = g.GetPlayPayout()
	resObj.TripsPayout = g.GetTripsPayout()
	resObj.TotalPayout = g.GetTotalPayout()
	resObj.PlayerHandRank = g.GetPlayerHandRank()
	resObj.DealerHandRank = g.GetDealerHandRank()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() {
		switch g.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "ultimatetexasholdem.result.playerWins"
		case domain.GameResultLose:
			if g.GetFolded() {
				resObj.Message = "Player folded."
				resObj.MessageCode = "ultimatetexasholdem.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "ultimatetexasholdem.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "ultimatetexasholdem.result.push"
		default:
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントを出力する。Web ではヒントはクライアント側 (useGameHint) で
// 算出するため、通常の状態出力を返す。UltimateTexasHoldemPresenter インタフェースを
// 満たすための実装。
func (up *UltimateTexasHoldemWebPresenter) HintOutput(g interfaces.UltimateTexasHoldemGame) string {
	return up.Output(g, nil)
}

// ActionLogOutput 棋譜をJSON出力
func (up *UltimateTexasHoldemWebPresenter) ActionLogOutput(g interfaces.UltimateTexasHoldemGame) string {
	return actionLogOutputJSON(g)
}

// ultimateTexasHoldemMaskDealerHand returns the dealer hand with all cards masked
// while the showdown has not yet happened.
func ultimateTexasHoldemMaskDealerHand(cards []*domain.Card) []*controller.WebOutputCard {
	if len(cards) == 0 {
		return make([]*controller.WebOutputCard, 0)
	}
	result := make([]*controller.WebOutputCard, len(cards))
	for i := range cards {
		result[i] = &controller.WebOutputCard{Design: "", Value: 0}
	}
	return result
}
