//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SimpleSimonWebInput シンプル・サイモンのWebインプット。
type SimpleSimonWebInput struct {
	BaseWebInput
	// FromCol 移動元の列番号。
	FromCol *int `json:"fromCol,omitempty"`
	// CardIndex 移動を開始するカードの列内インデックス。
	CardIndex *int `json:"cardIndex,omitempty"`
	// ToCol 移動先の列番号。
	ToCol *int `json:"toCol,omitempty"`
}

// SimpleSimonWebOutputHint ヒント出力。
type SimpleSimonWebOutputHint struct {
	FromCol   int `json:"fromCol"`
	CardIndex int `json:"cardIndex"`
	ToCol     int `json:"toCol"`
}

// SimpleSimonWebOutput シンプル・サイモンのWebアウトプット。
type SimpleSimonWebOutput struct {
	Columns        [][]*WebOutputCard        `json:"columns"`
	CompletedSuits int                       `json:"completedSuits"`
	Phase          int                       `json:"phase"`
	MoveCount      int                       `json:"moveCount"`
	CanUndo        bool                      `json:"canUndo"`
	Hint           *SimpleSimonWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// SimpleSimonWebController シンプル・サイモンのWebコントローラークラス。
type SimpleSimonWebController = GameWebController[usecase.SimpleSimonInteractorIF, SimpleSimonWebInput, *SimpleSimonWebOutput]

// NewSimpleSimonWebController and NewSimpleSimonWebControllerWithProvider are the
// standard and provider-backed constructors for SimpleSimonWebController.
var NewSimpleSimonWebController, NewSimpleSimonWebControllerWithProvider = webControllerPair[usecase.SimpleSimonInteractorIF, SimpleSimonWebInput, *SimpleSimonWebOutput](
	newSimpleSimonDefaultOutput, simpleSimonDispatch,
)

func newSimpleSimonDefaultOutput(msg string) *SimpleSimonWebOutput {
	return &SimpleSimonWebOutput{
		Columns:       make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func simpleSimonDispatch(bc *baseController, w http.ResponseWriter, si usecase.SimpleSimonInteractorIF, param SimpleSimonWebInput, newDefault func(string) *SimpleSimonWebOutput) bool {
	switch param.Command {
	case "m", "move":
		if !requireParam(bc, w, newDefault, param.FromCol == nil || param.CardIndex == nil || param.ToCol == nil, "param error: fromCol, cardIndex and toCol are required.") {
			return true
		}
		bc.writePresenterResponse(w, si.MoveSequence(*param.FromCol, *param.CardIndex, *param.ToCol))
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
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
