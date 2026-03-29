package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"net/http"
)

// ShortDeckWebInput ショートデックホールデムWebインプット (HoldemWebInputと同一構造)
type ShortDeckWebInput = HoldemWebInput

// ShortDeckWebOutput ショートデックホールデムWebアウトプット (HoldemWebOutputと同一構造)
type ShortDeckWebOutput = HoldemWebOutput

// ShortDeckWebController ショートデックホールデムWebコントローラークラス
type ShortDeckWebController = GameWebController[usecase.ShortDeckInteractorIF, ShortDeckWebInput, *ShortDeckWebOutput]

// NewShortDeckWebController コンストラクタ
func NewShortDeckWebController(factory func() usecase.ShortDeckInteractorIF) *ShortDeckWebController {
	return NewGameWebController(factory, newShortDeckDefaultOutput, shortDeckDispatch)
}

// NewShortDeckWebControllerWithProvider creates a ShortDeckWebController with an
// explicit SessionProvider (e.g. KV-backed for Workers).
func NewShortDeckWebControllerWithProvider(
	provider SessionProvider[usecase.ShortDeckInteractorIF],
	factory func() usecase.ShortDeckInteractorIF,
) *ShortDeckWebController {
	return NewGameWebControllerWithProvider(provider, factory, newShortDeckDefaultOutput, shortDeckDispatch)
}

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
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, 400, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, ogi.ResetWithConfig(cfg, param.Profile))
	case "f", "fold":
		bc.writePresenterResponse(w, ogi.Action(domain.ShortDeckActionFold, 0, param.HumanPlayMs))
	case "ck", "check":
		bc.writePresenterResponse(w, ogi.Action(domain.ShortDeckActionCheck, 0, param.HumanPlayMs))
	case "c", "call":
		bc.writePresenterResponse(w, ogi.Action(domain.ShortDeckActionCall, 0, param.HumanPlayMs))
	case "b", "bet":
		bc.writePresenterResponse(w, ogi.Action(domain.ShortDeckActionBet, param.Amount, param.HumanPlayMs))
	case "ra", "raise":
		bc.writePresenterResponse(w, ogi.Action(domain.ShortDeckActionRaise, param.Amount, param.HumanPlayMs))
	case "a", "allin":
		bc.writePresenterResponse(w, ogi.Action(domain.ShortDeckActionAllIn, 0, param.HumanPlayMs))
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
