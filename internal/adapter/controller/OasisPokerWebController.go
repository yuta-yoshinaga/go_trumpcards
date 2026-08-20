//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OasisPokerWebInput オアシスポーカーWebインプット
type OasisPokerWebInput struct {
	BaseWebInput
	Amount     int   `json:"amount,omitempty"`
	JackpotBet *int  `json:"jackpotBet,omitempty"`
	Indices    []int `json:"indices,omitempty"`
}

// OasisPokerWebOutput オアシスポーカーWebアウトプット
type OasisPokerWebOutput struct {
	PlayerHand      []*WebOutputCard `json:"playerHand"`
	DealerHand      []*WebOutputCard `json:"dealerHand"`
	Phase           int              `json:"phase"`
	Chips           int              `json:"chips"`
	AnteBet         int              `json:"anteBet"`
	JackpotBet      int              `json:"jackpotBet"`
	ExchangeCount   int              `json:"exchangeCount"`
	ExchangeFee     int              `json:"exchangeFee"`
	PlayBet         int              `json:"playBet"`
	Result          int              `json:"result"`
	AntePayout      int              `json:"antePayout"`
	PlayPayout      int              `json:"playPayout"`
	JackpotPayout   int              `json:"jackpotPayout"`
	TotalPayout     int              `json:"totalPayout"`
	DealerQualified bool             `json:"dealerQualified"`
	PlayerHandRank  int              `json:"playerHandRank"`
	DealerHandRank  int              `json:"dealerHandRank"`
	WebOutputBase
}

// OasisPokerWebController オアシスポーカーWebコントローラークラス
type OasisPokerWebController = GameWebController[usecase.OasisPokerInteractorIF, OasisPokerWebInput, *OasisPokerWebOutput]

// NewOasisPokerWebController and NewOasisPokerWebControllerWithProvider are
// the standard and provider-backed constructors for OasisPokerWebController.
var NewOasisPokerWebController, NewOasisPokerWebControllerWithProvider = webControllerPair[usecase.OasisPokerInteractorIF, OasisPokerWebInput, *OasisPokerWebOutput](
	newOasisPokerDefaultOutput, oasisPokerDispatch,
)

func newOasisPokerDefaultOutput(msg string) *OasisPokerWebOutput {
	return &OasisPokerWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func oasisPokerDispatch(bc *baseController, w http.ResponseWriter, oi usecase.OasisPokerInteractorIF, param OasisPokerWebInput, _ func(string) *OasisPokerWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		jackpot := deref(param.JackpotBet)
		bc.writePresenterResponse(w, oi.Bet(param.Amount, jackpot))
	case "e", "exchange":
		bc.writePresenterResponse(w, oi.Exchange(param.Indices))
	case "s", "stand":
		bc.writePresenterResponse(w, oi.Stand())
	case "p", "play":
		bc.writePresenterResponse(w, oi.Play())
	case "f", "fold":
		bc.writePresenterResponse(w, oi.Fold())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, oi.Reset, oi.Hint, oi.ActionLog)
	}
	return true
}
