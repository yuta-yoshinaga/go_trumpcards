//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PaiGowWebInput パイガオポーカーWebインプット
type PaiGowWebInput struct {
	BaseWebInput
	Amount int  `json:"amount,omitempty"`
	Low0   *int `json:"low0,omitempty"`
	Low1   *int `json:"low1,omitempty"`
}

// PaiGowWebOutput パイガオポーカーWebアウトプット
type PaiGowWebOutput struct {
	PlayerCards    []*WebOutputCard     `json:"playerCards"`
	DealerCards    []*WebOutputCard     `json:"dealerCards"`
	PlayerHighHand []*WebOutputCard     `json:"playerHighHand"`
	PlayerLowHand  []*WebOutputCard     `json:"playerLowHand"`
	DealerHighHand []*WebOutputCard     `json:"dealerHighHand"`
	DealerLowHand  []*WebOutputCard     `json:"dealerLowHand"`
	Phase          int                  `json:"phase"`
	Chips          int                  `json:"chips"`
	Bet            int                  `json:"bet"`
	Result         int                  `json:"result"`
	HighHandResult int                  `json:"highHandResult"`
	LowHandResult  int                  `json:"lowHandResult"`
	Payout         int                  `json:"payout"`
	Commission     int                  `json:"commission"`
	PlayerHighRank int                  `json:"playerHighRank"`
	PlayerLowRank  int                  `json:"playerLowRank"`
	DealerHighRank int                  `json:"dealerHighRank"`
	DealerLowRank  int                  `json:"dealerLowRank"`
	Hint           *PaiGowWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// PaiGowWebOutputHint はセットハンドフェーズの推奨分割。
type PaiGowWebOutputHint struct {
	LowIdx0   int    `json:"lowIdx0"`
	LowIdx1   int    `json:"lowIdx1"`
	LowIsPair bool   `json:"lowIsPair"`
	Reason    string `json:"reason"`
}

// PaiGowWebController パイガオポーカーWebコントローラークラス
type PaiGowWebController = GameWebController[usecase.PaiGowInteractorIF, PaiGowWebInput, *PaiGowWebOutput]

// NewPaiGowWebController and NewPaiGowWebControllerWithProvider are
// the standard and provider-backed constructors for PaiGowWebController.
var NewPaiGowWebController, NewPaiGowWebControllerWithProvider = webControllerPair[usecase.PaiGowInteractorIF, PaiGowWebInput, *PaiGowWebOutput](
	newPaiGowDefaultOutput, paiGowDispatch,
)

func newPaiGowDefaultOutput(msg string) *PaiGowWebOutput {
	return &PaiGowWebOutput{
		PlayerCards:    make([]*WebOutputCard, 0),
		DealerCards:    make([]*WebOutputCard, 0),
		PlayerHighHand: make([]*WebOutputCard, 0),
		PlayerLowHand:  make([]*WebOutputCard, 0),
		DealerHighHand: make([]*WebOutputCard, 0),
		DealerLowHand:  make([]*WebOutputCard, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func paiGowDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PaiGowInteractorIF, param PaiGowWebInput, _ func(string) *PaiGowWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, pi.Bet(param.Amount))
	case "s", "set":
		low0 := deref(param.Low0)
		low1 := deref(param.Low1)
		bc.writePresenterResponse(w, pi.SetHands(low0, low1))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, pi.Reset, pi.Hint, pi.ActionLog)
	}
	return true
}
