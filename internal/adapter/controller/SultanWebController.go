//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SultanWebInput スルタンWebインプット
type SultanWebInput struct {
	BaseWebInput
	From *SultanWebZone `json:"from,omitempty"`
	To   *SultanWebZone `json:"to,omitempty"`
}

// SultanWebZone ゾーン指定
type SultanWebZone struct {
	Zone     string `json:"zone"`
	DivanIdx *int   `json:"divanIdx,omitempty"`
}

// SultanWebOutputHint ヒント出力
type SultanWebOutputHint struct {
	FromZone     string `json:"fromZone"`
	FromIdx      int    `json:"fromIdx"`
	ToFoundation int    `json:"toFoundation"`
}

// SultanWebOutput スルタンWebアウトプット
type SultanWebOutput struct {
	Foundation  [][]*WebOutputCard   `json:"foundation"`
	Divan       []*WebOutputCard     `json:"divan"`
	StockCount  int                  `json:"stockCount"`
	Waste       []*WebOutputCard     `json:"waste"`
	RedealCount int                  `json:"redealCount"`
	CanRedeal   bool                 `json:"canRedeal"`
	Hint        *SultanWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// SultanWebController スルタンWebコントローラークラス
type SultanWebController = GameWebController[usecase.SultanInteractorIF, SultanWebInput, *SultanWebOutput]

// NewSultanWebController and NewSultanWebControllerWithProvider are the standard
// and provider-backed constructors for SultanWebController.
var NewSultanWebController, NewSultanWebControllerWithProvider = webControllerPair[usecase.SultanInteractorIF, SultanWebInput, *SultanWebOutput](
	newSultanDefaultOutput, sultanDispatch,
)

func newSultanDefaultOutput(msg string) *SultanWebOutput {
	return &SultanWebOutput{
		Foundation:    make([][]*WebOutputCard, 0),
		Divan:         make([]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func sultanDispatch(bc *baseController, w http.ResponseWriter, si usecase.SultanInteractorIF, param SultanWebInput, newDefault func(string) *SultanWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, si.Draw())
	case "rd", "redeal":
		bc.writePresenterResponse(w, si.Redeal())
	case "m", "move":
		return sultanMoveDispatch(bc, w, si, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, si.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, si.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, si.Reset, si.Hint, si.ActionLog)
	}
	return true
}

func sultanMoveDispatch(bc *baseController, w http.ResponseWriter, si usecase.SultanInteractorIF, param SultanWebInput, newDefault func(string) *SultanWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil, "param error: from is required.") {
		return true
	}
	switch param.From.Zone {
	case "divan":
		if !requireParam(bc, w, newDefault, param.From.DivanIdx == nil, "param error: from.divanIdx is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.MoveDivanToFoundation(*param.From.DivanIdx))
	case "waste":
		bc.writePresenterResponse(w, si.MoveWasteToFoundation())
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
