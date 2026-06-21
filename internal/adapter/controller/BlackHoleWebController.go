//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BlackHoleWebInput ブラックホールのWebインプット。
type BlackHoleWebInput struct {
	BaseWebInput
	// Fan 移動元の扇番号。
	Fan *int `json:"fan,omitempty"`
}

// BlackHoleWebOutputHint ヒント出力。
type BlackHoleWebOutputHint struct {
	Fan int `json:"fan"`
}

// BlackHoleWebOutput ブラックホールのWebアウトプット。
type BlackHoleWebOutput struct {
	Fans        [][]*WebOutputCard      `json:"fans"`
	BlackHole   []*WebOutputCard        `json:"blackHole"`
	Phase       int                     `json:"phase"`
	MoveCount   int                     `json:"moveCount"`
	CanUndo     bool                    `json:"canUndo"`
	IsStalemate bool                    `json:"isStalemate"`
	Hint        *BlackHoleWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// BlackHoleWebController ブラックホールのWebコントローラークラス。
type BlackHoleWebController = GameWebController[usecase.BlackHoleInteractorIF, BlackHoleWebInput, *BlackHoleWebOutput]

// NewBlackHoleWebController and NewBlackHoleWebControllerWithProvider are the
// standard and provider-backed constructors for BlackHoleWebController.
var NewBlackHoleWebController, NewBlackHoleWebControllerWithProvider = webControllerPair[usecase.BlackHoleInteractorIF, BlackHoleWebInput, *BlackHoleWebOutput](
	newBlackHoleDefaultOutput, blackHoleDispatch,
)

func newBlackHoleDefaultOutput(msg string) *BlackHoleWebOutput {
	return &BlackHoleWebOutput{
		Fans:          make([][]*WebOutputCard, 0),
		BlackHole:     make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func blackHoleDispatch(bc *baseController, w http.ResponseWriter, li usecase.BlackHoleInteractorIF, param BlackHoleWebInput, newDefault func(string) *BlackHoleWebOutput) bool {
	switch param.Command {
	case "mb", "m":
		if !requireParam(bc, w, newDefault, param.Fan == nil, "param error: fan is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.MoveFanToBlackHole(*param.Fan))
	case "g", "giveup":
		bc.writePresenterResponse(w, li.GiveUp())
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
