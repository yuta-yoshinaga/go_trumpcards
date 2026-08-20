//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DragonTigerWebInput ドラゴンタイガーWebインプット
type DragonTigerWebInput struct {
	BaseWebInput
	Amount  int `json:"amount,omitempty"`
	BetType int `json:"betType,omitempty"`
}

// DragonTigerWebOutput ドラゴンタイガーWebアウトプット
type DragonTigerWebOutput struct {
	DragonCard *WebOutputCard `json:"dragonCard,omitempty"`
	TigerCard  *WebOutputCard `json:"tigerCard,omitempty"`
	Phase      int            `json:"phase"`
	Chips      int            `json:"chips"`
	BetAmount  int            `json:"betAmount"`
	BetType    int            `json:"betType"`
	Result     int            `json:"result"`
	Payout     int            `json:"payout"`
	History    []int          `json:"history"`
	WebOutputBase
}

// DragonTigerWebController ドラゴンタイガーWebコントローラークラス
type DragonTigerWebController = GameWebController[usecase.DragonTigerInteractorIF, DragonTigerWebInput, *DragonTigerWebOutput]

// NewDragonTigerWebController and NewDragonTigerWebControllerWithProvider are
// the standard and provider-backed constructors for DragonTigerWebController.
var NewDragonTigerWebController, NewDragonTigerWebControllerWithProvider = webControllerPair[usecase.DragonTigerInteractorIF, DragonTigerWebInput, *DragonTigerWebOutput](
	newDragonTigerDefaultOutput, dragonTigerDispatch,
)

func newDragonTigerDefaultOutput(msg string) *DragonTigerWebOutput {
	return &DragonTigerWebOutput{
		History:       make([]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func dragonTigerDispatch(bc *baseController, w http.ResponseWriter, di usecase.DragonTigerInteractorIF, param DragonTigerWebInput, _ func(string) *DragonTigerWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, di.Bet(param.Amount, param.BetType))
	case "clear":
		bc.writePresenterResponse(w, di.ClearHistory())
	default:
		return dispatchResetAndLog(param.Command, bc, w, di.Reset, di.ActionLog)
	}
	return true
}
