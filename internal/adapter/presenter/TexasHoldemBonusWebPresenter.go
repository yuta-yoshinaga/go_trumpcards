package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TexasHoldemBonusWebPresenter テキサスホールデムボーナスポーカーWebプレゼンタークラス
type TexasHoldemBonusWebPresenter struct{}

// Output ゲーム状態を出力
func (tp *TexasHoldemBonusWebPresenter) Output(g interfaces.TexasHoldemBonusGame, lastErr error) string {
	resObj := new(controller.TexasHoldemBonusWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(g.GetPlayerHand())
	if g.GetPhase() == domain.TexasHoldemBonusPhaseEnd {
		resObj.DealerHand = cardsToOutputOrEmpty(g.GetDealerHand())
	} else {
		resObj.DealerHand = texasHoldemBonusMaskDealerHand(g.GetDealerHand())
	}
	resObj.Community = cardsToOutputOrEmpty(g.GetCommunity())
	resObj.Phase = g.GetPhase()
	resObj.Chips = g.GetChips()
	resObj.AnteBet = g.GetAnteBet()
	resObj.BonusBet = g.GetBonusBet()
	resObj.FlopBet = g.GetFlopBet()
	resObj.TurnBet = g.GetTurnBet()
	resObj.RiverBet = g.GetRiverBet()
	resObj.TotalPlayBet = g.GetTotalPlayBet()
	resObj.Result = int(g.GetResult())
	resObj.AntePayout = g.GetAntePayout()
	resObj.PlayPayout = g.GetPlayPayout()
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
			resObj.MessageCode = "texasholdembonus.result.playerWins"
		case domain.GameResultLose:
			if g.GetTotalPlayBet() == 0 {
				resObj.Message = "Player folded."
				resObj.MessageCode = "texasholdembonus.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "texasholdembonus.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "texasholdembonus.result.push"
		default:
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (tp *TexasHoldemBonusWebPresenter) ActionLogOutput(g interfaces.TexasHoldemBonusGame) string {
	return actionLogOutputJSON(g)
}

// texasHoldemBonusMaskDealerHand returns the dealer hand with all cards masked
// while the showdown has not yet happened.
func texasHoldemBonusMaskDealerHand(cards []*domain.Card) []*controller.WebOutputCard {
	if len(cards) == 0 {
		return make([]*controller.WebOutputCard, 0)
	}
	result := make([]*controller.WebOutputCard, len(cards))
	for i := range cards {
		result[i] = &controller.WebOutputCard{Design: "", Value: 0}
	}
	return result
}
