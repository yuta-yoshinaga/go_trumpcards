//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RedDogWebInput レッドドッグWebインプット
type RedDogWebInput struct {
	BaseWebInput
	Amount int `json:"amount,omitempty"`
}

// RedDogWebOutput レッドドッグWebアウトプット
type RedDogWebOutput struct {
	InitialCards []*WebOutputCard `json:"initialCards"`
	ThirdCard    *WebOutputCard   `json:"thirdCard,omitempty"`
	Phase        int              `json:"phase"`
	Chips        int              `json:"chips"`
	Ante         int              `json:"ante"`
	Raise        int              `json:"raise"`
	Spread       int              `json:"spread"`
	Result       int              `json:"result"`
	TotalPayout  int              `json:"totalPayout"`
	WebOutputBase
}

// RedDogWebController レッドドッグWebコントローラークラス
type RedDogWebController = GameWebController[usecase.RedDogInteractorIF, RedDogWebInput, *RedDogWebOutput]

// NewRedDogWebController and NewRedDogWebControllerWithProvider are
// the standard and provider-backed constructors for RedDogWebController.
var NewRedDogWebController, NewRedDogWebControllerWithProvider = webControllerPair[usecase.RedDogInteractorIF, RedDogWebInput, *RedDogWebOutput](
	newRedDogDefaultOutput, redDogDispatch,
)

func newRedDogDefaultOutput(msg string) *RedDogWebOutput {
	return &RedDogWebOutput{
		InitialCards:  make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func redDogDispatch(bc *baseController, w http.ResponseWriter, ri usecase.RedDogInteractorIF, param RedDogWebInput, _ func(string) *RedDogWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, ri.Bet(param.Amount))
	case "raise":
		bc.writePresenterResponse(w, ri.Raise(param.Amount))
	case "s", "stay":
		bc.writePresenterResponse(w, ri.Stay())
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ri.Reset, ri.Hint, ri.ActionLog)
	}
	return true
}
