//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CurdsAndWheyWebInput カーズ・アンド・ホエイのWebインプット。
type CurdsAndWheyWebInput struct {
	BaseWebInput
	// FromCol 移動元の列番号。
	FromCol *int `json:"fromCol,omitempty"`
	// CardIndex 移動を開始するカードの列内インデックス。
	CardIndex *int `json:"cardIndex,omitempty"`
	// ToCol 移動先の列番号。
	ToCol *int `json:"toCol,omitempty"`
}

// CurdsAndWheyWebOutputHint ヒント出力。
type CurdsAndWheyWebOutputHint struct {
	FromCol   int `json:"fromCol"`
	CardIndex int `json:"cardIndex"`
	ToCol     int `json:"toCol"`
}

// CurdsAndWheyWebOutput カーズ・アンド・ホエイのWebアウトプット。
type CurdsAndWheyWebOutput struct {
	Columns        [][]*WebOutputCard         `json:"columns"`
	CompletedSuits int                        `json:"completedSuits"`
	Phase          int                        `json:"phase"`
	MoveCount      int                        `json:"moveCount"`
	CanUndo        bool                       `json:"canUndo"`
	Hint           *CurdsAndWheyWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// CurdsAndWheyWebController カーズ・アンド・ホエイのWebコントローラークラス。
type CurdsAndWheyWebController = GameWebController[usecase.CurdsAndWheyInteractorIF, CurdsAndWheyWebInput, *CurdsAndWheyWebOutput]

// NewCurdsAndWheyWebController and NewCurdsAndWheyWebControllerWithProvider are the
// standard and provider-backed constructors for CurdsAndWheyWebController.
var NewCurdsAndWheyWebController, NewCurdsAndWheyWebControllerWithProvider = webControllerPair[usecase.CurdsAndWheyInteractorIF, CurdsAndWheyWebInput, *CurdsAndWheyWebOutput](
	newCurdsAndWheyDefaultOutput, curdsAndWheyDispatch,
)

func newCurdsAndWheyDefaultOutput(msg string) *CurdsAndWheyWebOutput {
	return &CurdsAndWheyWebOutput{
		Columns:       make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func curdsAndWheyDispatch(bc *baseController, w http.ResponseWriter, si usecase.CurdsAndWheyInteractorIF, param CurdsAndWheyWebInput, newDefault func(string) *CurdsAndWheyWebOutput) bool {
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
