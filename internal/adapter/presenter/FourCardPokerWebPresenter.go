package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FourCardPokerWebPresenter is the Four Card Poker Web presenter.
type FourCardPokerWebPresenter struct{}

// Output renders the game state as JSON.
func (p *FourCardPokerWebPresenter) Output(g interfaces.FourCardPokerGame, lastErr error) string {
	resObj := new(controller.FourCardPokerWebOutput)

	resObj.PlayerHand = cardsToOutputOrEmpty(g.GetPlayerHand())
	// Hide the dealer's hole cards during the action phase — only the upcard is visible.
	if g.GetPhase() == domain.FourCardPokerPhaseAction {
		if up := g.GetDealerUpCard(); up != nil {
			resObj.DealerHand = cardsToOutputOrEmpty([]*domain.Card{up})
		} else {
			resObj.DealerHand = make([]*controller.WebOutputCard, 0)
		}
	} else {
		resObj.DealerHand = cardsToOutputOrEmpty(g.GetDealerHand())
	}
	resObj.PlayerBest = cardsToOutputOrEmpty(g.GetPlayerBest())
	if g.GetPhase() == domain.FourCardPokerPhaseEnd {
		resObj.DealerBest = cardsToOutputOrEmpty(g.GetDealerBest())
	} else {
		resObj.DealerBest = make([]*controller.WebOutputCard, 0)
	}
	resObj.Phase = g.GetPhase()
	resObj.Chips = g.GetChips()
	resObj.AnteBet = g.GetAnteBet()
	resObj.AcesUpBet = g.GetAcesUpBet()
	resObj.PlayBet = g.GetPlayBet()
	resObj.PlayMultiplier = g.GetPlayMultiplier()
	resObj.Result = int(g.GetResult())
	resObj.AntePayout = g.GetAntePayout()
	resObj.PlayPayout = g.GetPlayPayout()
	resObj.AnteBonusPayout = g.GetAnteBonusPayout()
	resObj.AcesUpPayout = g.GetAcesUpPayout()
	resObj.TotalPayout = g.GetTotalPayout()
	resObj.PlayerHandRank = g.GetPlayerHandRank()
	resObj.DealerHandRank = g.GetDealerHandRank()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() {
		switch g.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "fourcardpoker.result.playerWins"
		case domain.GameResultLose:
			if g.GetPlayBet() == 0 {
				resObj.Message = "Player folded."
				resObj.MessageCode = "fourcardpoker.result.fold"
			} else {
				resObj.Message = "Dealer wins!"
				resObj.MessageCode = "fourcardpoker.result.dealerWins"
			}
		case domain.GameResultDraw:
			resObj.Message = "Push!"
			resObj.MessageCode = "fourcardpoker.result.push"
		default:
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput returns the action log as JSON.
func (p *FourCardPokerWebPresenter) ActionLogOutput(g interfaces.FourCardPokerGame) string {
	return actionLogOutputJSON(g)
}
