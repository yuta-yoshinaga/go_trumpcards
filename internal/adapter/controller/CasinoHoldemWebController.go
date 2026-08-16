//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CasinoHoldemWebInput カジノホールデムWebインプット
type CasinoHoldemWebInput struct {
	BaseWebInput
	Amount   int  `json:"amount,omitempty"`
	BonusBet *int `json:"bonusBet,omitempty"`
}

// CasinoHoldemWebOutput カジノホールデムWebアウトプット
type CasinoHoldemWebOutput struct {
	PlayerHand     []*WebOutputCard `json:"playerHand"`
	DealerHand     []*WebOutputCard `json:"dealerHand"`
	Community      []*WebOutputCard `json:"community"`
	Phase          int              `json:"phase"`
	Chips          int              `json:"chips"`
	AnteBet        int              `json:"anteBet"`
	BonusBet       int              `json:"bonusBet"`
	CallBet        int              `json:"callBet"`
	Result         int              `json:"result"`
	DealerQualify  bool             `json:"dealerQualify"`
	AntePayout     int              `json:"antePayout"`
	CallPayout     int              `json:"callPayout"`
	BonusPayout    int              `json:"bonusPayout"`
	TotalPayout    int              `json:"totalPayout"`
	PlayerHandRank int              `json:"playerHandRank"`
	DealerHandRank int              `json:"dealerHandRank"`
	WebOutputBase
}

// CasinoHoldemWebController カジノホールデムWebコントローラークラス
type CasinoHoldemWebController = GameWebController[usecase.CasinoHoldemInteractorIF, CasinoHoldemWebInput, *CasinoHoldemWebOutput]

// NewCasinoHoldemWebController and NewCasinoHoldemWebControllerWithProvider are
// the standard and provider-backed constructors for CasinoHoldemWebController.
var NewCasinoHoldemWebController, NewCasinoHoldemWebControllerWithProvider = webControllerPair[usecase.CasinoHoldemInteractorIF, CasinoHoldemWebInput, *CasinoHoldemWebOutput](
	newCasinoHoldemDefaultOutput, casinoHoldemDispatch,
)

func newCasinoHoldemDefaultOutput(msg string) *CasinoHoldemWebOutput {
	return &CasinoHoldemWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		Community:     make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func casinoHoldemDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CasinoHoldemInteractorIF, param CasinoHoldemWebInput, _ func(string) *CasinoHoldemWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bonus := deref(param.BonusBet)
		bc.writePresenterResponse(w, ci.Bet(param.Amount, bonus))
	case "c", "call":
		bc.writePresenterResponse(w, ci.Call())
	case "f", "fold":
		bc.writePresenterResponse(w, ci.Fold())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ci.Reset, ci.Hint, ci.ActionLog)
	}
	return true
}
