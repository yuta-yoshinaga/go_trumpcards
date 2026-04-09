package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ShortDeckWebInput ショートデックホールデムWebインプット (HoldemWebInputと同一構造)
type ShortDeckWebInput = HoldemWebInput

// ShortDeckWebOutput ショートデックホールデムWebアウトプット (HoldemWebOutputと同一構造)
type ShortDeckWebOutput = HoldemWebOutput

// ShortDeckWebController ショートデックホールデムWebコントローラークラス
type ShortDeckWebController = GameWebController[usecase.ShortDeckInteractorIF, ShortDeckWebInput, *ShortDeckWebOutput]

// NewShortDeckWebController and NewShortDeckWebControllerWithProvider are
// the standard and provider-backed constructors for ShortDeckWebController.
var NewShortDeckWebController, NewShortDeckWebControllerWithProvider = webControllerPair[usecase.ShortDeckInteractorIF, ShortDeckWebInput, *ShortDeckWebOutput](
	newShortDeckDefaultOutput, shortDeckDispatch,
)

func newShortDeckDefaultOutput(msg string) *ShortDeckWebOutput {
	return &ShortDeckWebOutput{
		Players:        make([]*HoldemWebOutputPlayer, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		SidePots:       make([]*HoldemWebOutputSidePot, 0),
		RoundResults:   make([]*HoldemWebOutputResult, 0),
		CpuActions:     make([]*HoldemWebOutputCpuAction, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func shortDeckDispatch(bc *baseController, w http.ResponseWriter, ogi usecase.ShortDeckInteractorIF, param ShortDeckWebInput, newDefault func(string) *ShortDeckWebOutput) bool {
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
