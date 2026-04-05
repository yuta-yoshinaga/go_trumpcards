package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"net/http"
)

// OmahaWebInput オマハホールデムWebインプット (HoldemWebInputと同一構造)
type OmahaWebInput = HoldemWebInput

// OmahaWebOutput オマハホールデムWebアウトプット (HoldemWebOutputと同一構造)
type OmahaWebOutput = HoldemWebOutput

// OmahaWebController オマハホールデムWebコントローラークラス
type OmahaWebController = GameWebController[usecase.OmahaInteractorIF, OmahaWebInput, *OmahaWebOutput]

// NewOmahaWebController and NewOmahaWebControllerWithProvider are
// the standard and provider-backed constructors for OmahaWebController.
var NewOmahaWebController, NewOmahaWebControllerWithProvider = WebControllerPair[usecase.OmahaInteractorIF, OmahaWebInput, *OmahaWebOutput](
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
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, 400, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, ogi.ResetWithConfig(cfg, param.Profile))
	case "f", "fold":
		bc.writePresenterResponse(w, ogi.Action(domain.OmahaActionFold, 0, param.HumanPlayMs))
	case "ck", "check":
		bc.writePresenterResponse(w, ogi.Action(domain.OmahaActionCheck, 0, param.HumanPlayMs))
	case "c", "call":
		bc.writePresenterResponse(w, ogi.Action(domain.OmahaActionCall, 0, param.HumanPlayMs))
	case "b", "bet":
		bc.writePresenterResponse(w, ogi.Action(domain.OmahaActionBet, param.Amount, param.HumanPlayMs))
	case "ra", "raise":
		bc.writePresenterResponse(w, ogi.Action(domain.OmahaActionRaise, param.Amount, param.HumanPlayMs))
	case "a", "allin":
		bc.writePresenterResponse(w, ogi.Action(domain.OmahaActionAllIn, 0, param.HumanPlayMs))
	case "rb", "rebuy":
		bc.writePresenterResponse(w, ogi.Rebuy())
	case "sr", "skiprebuy":
		bc.writePresenterResponse(w, ogi.SkipRebuy())
	case "ad", "addon":
		bc.writePresenterResponse(w, ogi.Addon())
	case "sa", "skipaddon":
		bc.writePresenterResponse(w, ogi.SkipAddon())
	case "m", "muck":
		bc.writePresenterResponse(w, ogi.Muck())
	case "sh", "show":
		bc.writePresenterResponse(w, ogi.ShowHand())
	default:
		return dispatchLog(param.Command, bc, w, ogi.ActionLog)
	}
	return true
}
