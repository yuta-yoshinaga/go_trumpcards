package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RussianPokerWebPresenter ロシアンポーカーWebプレゼンタークラス
type RussianPokerWebPresenter struct {
}

// Output ゲーム状態を出力
func (rp *RussianPokerWebPresenter) Output(g interfaces.RussianPokerGame, lastErr error) string {
	resObj := new(controller.RussianPokerWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(g.GetPlayerHand())
	if g.GetPhase() == domain.RussianPokerPhaseEnd {
		resObj.DealerHand = cardsToOutputOrEmpty(g.GetDealerHand())
	} else {
		resObj.DealerHand = russianPokerMaskDealerHand(g.GetDealerHand())
	}
	resObj.Phase = g.GetPhase()
	resObj.Chips = g.GetChips()
	resObj.AnteBet = g.GetAnteBet()
	resObj.ExchangeCount = g.GetExchangeCount()
	resObj.ExchangeFee = g.GetExchangeFee()
	resObj.Bought6th = g.GetBought6th()
	resObj.Buy6thFee = g.GetBuy6thFee()
	resObj.ForceExchanged = g.GetForceExchanged()
	resObj.ForceExchangeFee = g.GetForceExchangeFee()
	resObj.PlayBet = g.GetPlayBet()
	resObj.Result = int(g.GetResult())
	resObj.AntePayout = g.GetAntePayout()
	resObj.PlayPayout = g.GetPlayPayout()
	resObj.TotalPayout = g.GetTotalPayout()
	resObj.DealerQualified = g.GetDealerQualified()
	resObj.PlayerHandRank = g.GetPlayerHandRank()
	resObj.DealerHandRank = g.GetDealerHandRank()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() {
		switch g.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "russianpoker.result.playerWins"
		case domain.GameResultLose:
			if g.GetPlayBet() == 0 {
				resObj.Message = "Player folded."
				resObj.MessageCode = "russianpoker.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "russianpoker.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "russianpoker.result.push"
		default:
		}
		if !g.GetDealerQualified() && g.GetPlayBet() > 0 {
			resObj.Message = "Dealer does not qualify!"
			resObj.MessageCode = "russianpoker.result.dealerNotQualified"
		}
	} else if g.GetPhase() == domain.RussianPokerPhaseForceQualify {
		resObj.Message = "Dealer does not qualify. Force exchange?"
		resObj.MessageCode = "russianpoker.forceQualifyGuide"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (rp *RussianPokerWebPresenter) ActionLogOutput(g interfaces.RussianPokerGame) string {
	return actionLogOutputJSON(g)
}

// russianPokerMaskDealerHand returns the dealer hand with all cards except the first masked.
func russianPokerMaskDealerHand(cards []*domain.Card) []*controller.WebOutputCard {
	if len(cards) == 0 {
		return make([]*controller.WebOutputCard, 0)
	}
	result := make([]*controller.WebOutputCard, len(cards))
	result[0] = cardToOutput(cards[0])
	for i := 1; i < len(cards); i++ {
		result[i] = &controller.WebOutputCard{Design: "", Value: 0}
	}
	return result
}
