//go:build !js || !wasm || extra5

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OichoKabuWebInput おいちょかぶWebインプット
type OichoKabuWebInput struct {
	BaseWebInput
	Amount int `json:"amount,omitempty"`
}

// OichoKabuWebOutput おいちょかぶWebアウトプット
type OichoKabuWebOutput struct {
	PlayerHand  []*WebOutputCard `json:"playerHand"`
	BankerHand  []*WebOutputCard `json:"bankerHand"`
	PlayerRank  int              `json:"playerRank"`
	BankerRank  int              `json:"bankerRank"`
	Phase       int              `json:"phase"`
	Chips       int              `json:"chips"`
	Bet         int              `json:"bet"`
	Result      int              `json:"result"`
	TotalPayout int              `json:"totalPayout"`
	WebOutputBase
}

// OichoKabuWebController おいちょかぶWebコントローラークラス
type OichoKabuWebController = GameWebController[usecase.OichoKabuInteractorIF, OichoKabuWebInput, *OichoKabuWebOutput]

// NewOichoKabuWebController and NewOichoKabuWebControllerWithProvider are the
// standard and provider-backed constructors for OichoKabuWebController.
var NewOichoKabuWebController, NewOichoKabuWebControllerWithProvider = webControllerPair[usecase.OichoKabuInteractorIF, OichoKabuWebInput, *OichoKabuWebOutput](
	newOichoKabuDefaultOutput, oichoKabuDispatch,
)

func newOichoKabuDefaultOutput(msg string) *OichoKabuWebOutput {
	return &OichoKabuWebOutput{
		PlayerHand:    make([]*WebOutputCard, 0),
		BankerHand:    make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func oichoKabuDispatch(bc *baseController, w http.ResponseWriter, oi usecase.OichoKabuInteractorIF, param OichoKabuWebInput, _ func(string) *OichoKabuWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, oi.Bet(param.Amount))
	case "d", "draw":
		bc.writePresenterResponse(w, oi.Draw())
	case "s", "stand":
		bc.writePresenterResponse(w, oi.Stand())
	default:
		return dispatchResetAndLog(param.Command, bc, w, oi.Reset, oi.ActionLog)
	}
	return true
}
