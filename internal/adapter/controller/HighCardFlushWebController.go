//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HighCardFlushWebInput ハイカードフラッシュWebインプット
type HighCardFlushWebInput struct {
	BaseWebInput
	Amount           int  `json:"amount,omitempty"`
	FlushBonusBet    *int `json:"flushBonusBet,omitempty"`
	StraightFlushBet *int `json:"straightFlushBet,omitempty"`
	Multiplier       *int `json:"multiplier,omitempty"`
}

// HighCardFlushWebOutput ハイカードフラッシュWebアウトプット
type HighCardFlushWebOutput struct {
	PlayerHand             []*WebOutputCard `json:"playerHand"`
	DealerHand             []*WebOutputCard `json:"dealerHand"`
	Phase                  int              `json:"phase"`
	Chips                  int              `json:"chips"`
	AnteBet                int              `json:"anteBet"`
	FlushBonusBet          int              `json:"flushBonusBet"`
	StraightFlushBet       int              `json:"straightFlushBet"`
	RaiseBet               int              `json:"raiseBet"`
	Result                 int              `json:"result"`
	AntePayout             int              `json:"antePayout"`
	RaisePayout            int              `json:"raisePayout"`
	FlushBonusPayout       int              `json:"flushBonusPayout"`
	StraightFlushPayout    int              `json:"straightFlushPayout"`
	TotalPayout            int              `json:"totalPayout"`
	DealerQualified        bool             `json:"dealerQualified"`
	PlayerFlushLen         int              `json:"playerFlushLen"`
	DealerFlushLen         int              `json:"dealerFlushLen"`
	PlayerStraightFlushLen int              `json:"playerStraightFlushLen"`
	MaxRaiseMultiplier     int              `json:"maxRaiseMultiplier"`
	WebOutputBase
}

// HighCardFlushWebController ハイカードフラッシュWebコントローラークラス
type HighCardFlushWebController = GameWebController[usecase.HighCardFlushInteractorIF, HighCardFlushWebInput, *HighCardFlushWebOutput]

// NewHighCardFlushWebController and NewHighCardFlushWebControllerWithProvider are
// the standard and provider-backed constructors for HighCardFlushWebController.
var NewHighCardFlushWebController, NewHighCardFlushWebControllerWithProvider = webControllerPair[usecase.HighCardFlushInteractorIF, HighCardFlushWebInput, *HighCardFlushWebOutput](
	newHighCardFlushDefaultOutput, highCardFlushDispatch,
)

func newHighCardFlushDefaultOutput(msg string) *HighCardFlushWebOutput {
	return &HighCardFlushWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		DealerHand:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func highCardFlushDispatch(bc *baseController, w http.ResponseWriter, hi usecase.HighCardFlushInteractorIF, param HighCardFlushWebInput, _ func(string) *HighCardFlushWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		fb := deref(param.FlushBonusBet)
		sf := deref(param.StraightFlushBet)
		bc.writePresenterResponse(w, hi.Bet(param.Amount, fb, sf))
	case "ra", "raise":
		mult := deref(param.Multiplier)
		bc.writePresenterResponse(w, hi.Raise(mult))
	case "f", "fold":
		bc.writePresenterResponse(w, hi.Fold())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, hi.Reset, hi.Hint, hi.ActionLog)
	}
	return true
}
