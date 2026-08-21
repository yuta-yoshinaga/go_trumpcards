//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ShamrocksWebInput シャムロックスのWebインプット。
type ShamrocksWebInput struct {
	BaseWebInput
	// From 移動元の扇番号。
	From *int `json:"from,omitempty"`
	// To 移動先の扇番号 (mf のとき)。
	To *int `json:"to,omitempty"`
}

// ShamrocksWebOutputHint ヒント出力。
type ShamrocksWebOutputHint struct {
	FromFan      int  `json:"fromFan"`
	ToFan        int  `json:"toFan"`
	ToFoundation bool `json:"toFoundation"`
}

// ShamrocksWebOutput シャムロックスのWebアウトプット。
type ShamrocksWebOutput struct {
	Fans        [][]*WebOutputCard      `json:"fans"`
	Foundation  [][]*WebOutputCard      `json:"foundation"`
	RedealsLeft int                     `json:"redealsLeft"`
	Phase       int                     `json:"phase"`
	MoveCount   int                     `json:"moveCount"`
	CanUndo     bool                    `json:"canUndo"`
	Hint        *ShamrocksWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// ShamrocksWebController シャムロックスのWebコントローラークラス。
type ShamrocksWebController = GameWebController[usecase.ShamrocksInteractorIF, ShamrocksWebInput, *ShamrocksWebOutput]

// NewShamrocksWebController and NewShamrocksWebControllerWithProvider are
// the standard and provider-backed constructors for ShamrocksWebController.
var NewShamrocksWebController, NewShamrocksWebControllerWithProvider = webControllerPair[usecase.ShamrocksInteractorIF, ShamrocksWebInput, *ShamrocksWebOutput](
	newShamrocksDefaultOutput, shamrocksDispatch,
)

func newShamrocksDefaultOutput(msg string) *ShamrocksWebOutput {
	return &ShamrocksWebOutput{
		Fans:          make([][]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func shamrocksDispatch(bc *baseController, w http.ResponseWriter, li usecase.ShamrocksInteractorIF, param ShamrocksWebInput, newDefault func(string) *ShamrocksWebOutput) bool {
	switch param.Command {
	case "mf":
		if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
			return true
		}
		bc.writePresenterResponse(w, li.MoveFanToFan(*param.From, *param.To))
	case "ff":
		if !requireParam(bc, w, newDefault, param.From == nil, "param error: from is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.MoveFanToFoundation(*param.From))
	case "rd", "redeal":
		bc.writePresenterResponse(w, li.Redeal())
	case "g", "giveup":
		bc.writePresenterResponse(w, li.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, li.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, li.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, li.Reset, li.Hint, li.ActionLog)
	}
	return true
}
