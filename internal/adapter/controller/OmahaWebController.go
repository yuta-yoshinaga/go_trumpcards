package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OmahaWebInput オマハホールデムWebインプット (HoldemWebInputと同一構造)
type OmahaWebInput = HoldemWebInput

// OmahaWebOutput オマハホールデムWebアウトプット (HoldemWebOutputと同一構造)
type OmahaWebOutput = HoldemWebOutput

// OmahaWebController オマハホールデムWebコントローラークラス
type OmahaWebController = GameWebController[usecase.OmahaInteractorIF, OmahaWebInput, *OmahaWebOutput]

// NewOmahaWebController and NewOmahaWebControllerWithProvider are
// the standard and provider-backed constructors for OmahaWebController.
var NewOmahaWebController, NewOmahaWebControllerWithProvider = webControllerPair[usecase.OmahaInteractorIF, OmahaWebInput, *OmahaWebOutput](
	newOmahaDefaultOutput, omahaDispatch,
)

func newOmahaDefaultOutput(msg string) *OmahaWebOutput {
	return &OmahaWebOutput{
		Players:        make([]*HoldemWebOutputPlayer, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		SidePots:       make([]*HoldemWebOutputSidePot, 0),
		RoundResults:   make([]*HoldemWebOutputResult, 0),
		CpuActions:     make([]*HoldemWebOutputCpuAction, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func omahaDispatch(bc *baseController, w http.ResponseWriter, ogi usecase.OmahaInteractorIF, param OmahaWebInput, newDefault func(string) *OmahaWebOutput) bool {
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
