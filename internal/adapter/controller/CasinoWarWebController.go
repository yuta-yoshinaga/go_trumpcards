//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CasinoWarWebInput カジノウォーWebインプット
type CasinoWarWebInput struct {
	BaseWebInput
	Amount int `json:"amount,omitempty"`
}

// CasinoWarWebOutput カジノウォーWebアウトプット
type CasinoWarWebOutput struct {
	PlayerCard    *WebOutputCard   `json:"playerCard,omitempty"`
	DealerCard    *WebOutputCard   `json:"dealerCard,omitempty"`
	PlayerWarCard *WebOutputCard   `json:"playerWarCard,omitempty"`
	DealerWarCard *WebOutputCard   `json:"dealerWarCard,omitempty"`
	BurnCards     []*WebOutputCard `json:"burnCards"`
	Phase         int              `json:"phase"`
	Chips         int              `json:"chips"`
	Ante          int              `json:"ante"`
	WarBet        int              `json:"warBet"`
	Result        int              `json:"result"`
	TotalPayout   int              `json:"totalPayout"`
	WebOutputBase
}

// CasinoWarWebController カジノウォーWebコントローラークラス
type CasinoWarWebController = GameWebController[usecase.CasinoWarInteractorIF, CasinoWarWebInput, *CasinoWarWebOutput]

// NewCasinoWarWebController and NewCasinoWarWebControllerWithProvider are
// the standard and provider-backed constructors for CasinoWarWebController.
var NewCasinoWarWebController, NewCasinoWarWebControllerWithProvider = webControllerPair[usecase.CasinoWarInteractorIF, CasinoWarWebInput, *CasinoWarWebOutput](
	newCasinoWarDefaultOutput, casinoWarDispatch,
)

func newCasinoWarDefaultOutput(msg string) *CasinoWarWebOutput {
	return &CasinoWarWebOutput{
		BurnCards:     make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func casinoWarDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CasinoWarInteractorIF, param CasinoWarWebInput, _ func(string) *CasinoWarWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, ci.Bet(param.Amount))
	case "surrender":
		bc.writePresenterResponse(w, ci.Surrender())
	case "war":
		bc.writePresenterResponse(w, ci.War())
	default:
		return dispatchResetAndLog(param.Command, bc, w, ci.Reset, ci.ActionLog)
	}
	return true
}
