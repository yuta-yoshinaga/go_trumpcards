//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FourCardPokerWebInput is the request body for the Four Card Poker web endpoint.
type FourCardPokerWebInput struct {
	BaseWebInput
	Amount         int  `json:"amount,omitempty"`
	AcesUpBet      *int `json:"acesUpBet,omitempty"`
	PlayMultiplier *int `json:"playMultiplier,omitempty"`
}

// FourCardPokerWebOutput is the response body.
type FourCardPokerWebOutput struct {
	PlayerHand      []*WebOutputCard `json:"playerHand"`
	DealerHand      []*WebOutputCard `json:"dealerHand"`
	PlayerBest      []*WebOutputCard `json:"playerBest"`
	DealerBest      []*WebOutputCard `json:"dealerBest"`
	Phase           int              `json:"phase"`
	Chips           int              `json:"chips"`
	AnteBet         int              `json:"anteBet"`
	AcesUpBet       int              `json:"acesUpBet"`
	PlayBet         int              `json:"playBet"`
	PlayMultiplier  int              `json:"playMultiplier"`
	Result          int              `json:"result"`
	AntePayout      int              `json:"antePayout"`
	PlayPayout      int              `json:"playPayout"`
	AnteBonusPayout int              `json:"anteBonusPayout"`
	AcesUpPayout    int              `json:"acesUpPayout"`
	TotalPayout     int              `json:"totalPayout"`
	PlayerHandRank  int              `json:"playerHandRank"`
	DealerHandRank  int              `json:"dealerHandRank"`
	WebOutputBase
}

// FourCardPokerWebController is the Four Card Poker web controller type alias.
type FourCardPokerWebController = GameWebController[usecase.FourCardPokerInteractorIF, FourCardPokerWebInput, *FourCardPokerWebOutput]

// NewFourCardPokerWebController and NewFourCardPokerWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewFourCardPokerWebController, NewFourCardPokerWebControllerWithProvider = webControllerPair[usecase.FourCardPokerInteractorIF, FourCardPokerWebInput, *FourCardPokerWebOutput](
	newFourCardPokerDefaultOutput, fourCardPokerDispatch,
)

func newFourCardPokerDefaultOutput(msg string) *FourCardPokerWebOutput {
	return &FourCardPokerWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		PlayerBest:    make([]*WebOutputCard, 0),
		DealerBest:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func fourCardPokerDispatch(bc *baseController, w http.ResponseWriter, ti usecase.FourCardPokerInteractorIF, param FourCardPokerWebInput, _ func(string) *FourCardPokerWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		acesUp := deref(param.AcesUpBet)
		bc.writePresenterResponse(w, ti.Bet(param.Amount, acesUp))
	case "p", "play":
		mul := 1
		if param.PlayMultiplier != nil {
			mul = *param.PlayMultiplier
		}
		bc.writePresenterResponse(w, ti.Play(mul))
	case "f", "fold":
		bc.writePresenterResponse(w, ti.Fold())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ti.Reset, ti.ActionLog)
	}
	return true
}
