package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RussianPokerWebInput ロシアンポーカーWebインプット
type RussianPokerWebInput struct {
	BaseWebInput
	Amount       int   `json:"amount,omitempty"`
	Indices      []int `json:"indices,omitempty"`
	DiscardIndex *int  `json:"discardIndex,omitempty"`
}

// RussianPokerWebOutput ロシアンポーカーWebアウトプット
type RussianPokerWebOutput struct {
	PlayerHand       []*WebOutputCard `json:"playerHand"`
	DealerHand       []*WebOutputCard `json:"dealerHand"`
	Phase            int              `json:"phase"`
	Chips            int              `json:"chips"`
	AnteBet          int              `json:"anteBet"`
	ExchangeCount    int              `json:"exchangeCount"`
	ExchangeFee      int              `json:"exchangeFee"`
	Bought6th        bool             `json:"bought6th"`
	Buy6thFee        int              `json:"buy6thFee"`
	ForceExchanged   bool             `json:"forceExchanged"`
	ForceExchangeFee int              `json:"forceExchangeFee"`
	PlayBet          int              `json:"playBet"`
	Result           int              `json:"result"`
	AntePayout       int              `json:"antePayout"`
	PlayPayout       int              `json:"playPayout"`
	TotalPayout      int              `json:"totalPayout"`
	DealerQualified  bool             `json:"dealerQualified"`
	PlayerHandRank   int              `json:"playerHandRank"`
	DealerHandRank   int              `json:"dealerHandRank"`
	WebOutputBase
}

// RussianPokerWebController ロシアンポーカーWebコントローラークラス
type RussianPokerWebController = GameWebController[usecase.RussianPokerInteractorIF, RussianPokerWebInput, *RussianPokerWebOutput]

// NewRussianPokerWebController and NewRussianPokerWebControllerWithProvider are
// the standard and provider-backed constructors for RussianPokerWebController.
var NewRussianPokerWebController, NewRussianPokerWebControllerWithProvider = webControllerPair[usecase.RussianPokerInteractorIF, RussianPokerWebInput, *RussianPokerWebOutput](
	newRussianPokerDefaultOutput, russianPokerDispatch,
)

func newRussianPokerDefaultOutput(msg string) *RussianPokerWebOutput {
	return &RussianPokerWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func russianPokerDispatch(bc *baseController, w http.ResponseWriter, ri usecase.RussianPokerInteractorIF, param RussianPokerWebInput, _ func(string) *RussianPokerWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, ri.Bet(param.Amount))
	case "e", "exchange":
		bc.writePresenterResponse(w, ri.Exchange(param.Indices))
	case "buy6th", "6":
		bc.writePresenterResponse(w, ri.Buy6th())
	case "select", "sel":
		bc.writePresenterResponse(w, ri.Select(deref(param.DiscardIndex)))
	case "p", "play":
		bc.writePresenterResponse(w, ri.Play())
	case "f", "fold":
		bc.writePresenterResponse(w, ri.Fold())
	case "force", "fe":
		bc.writePresenterResponse(w, ri.ForceExchange())
	case "decline", "d":
		bc.writePresenterResponse(w, ri.Decline())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ri.Reset, ri.ActionLog)
	}
	return true
}
