//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OmahaWebController オマハホールデムWebコントローラークラス
type OmahaWebController = GameWebController[usecase.OmahaInteractorIF, HoldemWebInput, *HoldemWebOutput]

// NewOmahaWebController and NewOmahaWebControllerWithProvider are
// the standard and provider-backed constructors for OmahaWebController.
var NewOmahaWebController, NewOmahaWebControllerWithProvider = webControllerPair[usecase.OmahaInteractorIF, HoldemWebInput, *HoldemWebOutput](
	newOmahaDefaultOutput, omahaDispatch,
)

func newOmahaDefaultOutput(msg string) *HoldemWebOutput {
	return &HoldemWebOutput{
		Players:        make([]*HoldemWebOutputPlayer, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		SidePots:       make([]*HoldemWebOutputSidePot, 0),
		RoundResults:   make([]*HoldemWebOutputResult, 0),
		CpuActions:     make([]*HoldemWebOutputCpuAction, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func omahaDispatch(bc *baseController, w http.ResponseWriter, ogi usecase.OmahaInteractorIF, param HoldemWebInput, newDefault func(string) *HoldemWebOutput) bool {
	if dispatchPokerAction(bc, w, ogi, param.Command, param.Amount, param.HumanPlayMs) {
		return true
	}
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, ogi.ResetWithConfig(cfg, param.Profile))
	default:
		return dispatchLog(param.Command, bc, w, ogi.ActionLog)
	}
	return true
}
