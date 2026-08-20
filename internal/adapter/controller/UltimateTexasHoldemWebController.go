//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// UltimateTexasHoldemWebInput アルティメット・テキサスホールデムWebインプット
type UltimateTexasHoldemWebInput struct {
	BaseWebInput
	Amount     int  `json:"amount,omitempty"`
	TripsBet   *int `json:"tripsBet,omitempty"`
	Multiplier int  `json:"multiplier,omitempty"`
}

// UltimateTexasHoldemWebOutput アルティメット・テキサスホールデムWebアウトプット
type UltimateTexasHoldemWebOutput struct {
	PlayerHand      []*WebOutputCard `json:"playerHand"`
	DealerHand      []*WebOutputCard `json:"dealerHand"`
	Community       []*WebOutputCard `json:"community"`
	Phase           int              `json:"phase"`
	Chips           int              `json:"chips"`
	AnteBet         int              `json:"anteBet"`
	BlindBet        int              `json:"blindBet"`
	TripsBet        int              `json:"tripsBet"`
	PlayBet         int              `json:"playBet"`
	Folded          bool             `json:"folded"`
	Result          int              `json:"result"`
	DealerQualified bool             `json:"dealerQualified"`
	AntePayout      int              `json:"antePayout"`
	BlindPayout     int              `json:"blindPayout"`
	PlayPayout      int              `json:"playPayout"`
	TripsPayout     int              `json:"tripsPayout"`
	TotalPayout     int              `json:"totalPayout"`
	PlayerHandRank  int              `json:"playerHandRank"`
	DealerHandRank  int              `json:"dealerHandRank"`
	WebOutputBase
}

// UltimateTexasHoldemWebController アルティメット・テキサスホールデムWebコントローラー
type UltimateTexasHoldemWebController = GameWebController[usecase.UltimateTexasHoldemInteractorIF, UltimateTexasHoldemWebInput, *UltimateTexasHoldemWebOutput]

// NewUltimateTexasHoldemWebController and NewUltimateTexasHoldemWebControllerWithProvider are
// the standard and provider-backed constructors.
var NewUltimateTexasHoldemWebController, NewUltimateTexasHoldemWebControllerWithProvider = webControllerPair[usecase.UltimateTexasHoldemInteractorIF, UltimateTexasHoldemWebInput, *UltimateTexasHoldemWebOutput](
	newUltimateTexasHoldemDefaultOutput, ultimateTexasHoldemDispatch,
)

func newUltimateTexasHoldemDefaultOutput(msg string) *UltimateTexasHoldemWebOutput {
	return &UltimateTexasHoldemWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		Community:     make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func ultimateTexasHoldemDispatch(bc *baseController, w http.ResponseWriter, ui usecase.UltimateTexasHoldemInteractorIF, param UltimateTexasHoldemWebInput, _ func(string) *UltimateTexasHoldemWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		trips := deref(param.TripsBet)
		bc.writePresenterResponse(w, ui.Bet(param.Amount, trips))
	case "p", "play":
		bc.writePresenterResponse(w, ui.Play(param.Multiplier))
	case "c", "check":
		bc.writePresenterResponse(w, ui.Check())
	case "f", "fold":
		bc.writePresenterResponse(w, ui.Fold())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ui.Reset, ui.Hint, ui.ActionLog)
	}
	return true
}
