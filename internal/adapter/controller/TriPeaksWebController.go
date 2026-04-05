package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TriPeaksWebInput トリピークスWebインプット
type TriPeaksWebInput struct {
	BaseWebInput
	Row *int `json:"row,omitempty"`
	Col *int `json:"col,omitempty"`
}

// TriPeaksWebOutputCard タブローカード出力
type TriPeaksWebOutputCard struct {
	Card    *WebOutputCard `json:"card"`
	Removed bool           `json:"removed"`
	Exposed bool           `json:"exposed"`
}

// TriPeaksWebOutputHint ヒント出力
type TriPeaksWebOutputHint struct {
	Type string `json:"type"`
	Row  int    `json:"row"`
	Col  int    `json:"col"`
}

// TriPeaksWebOutput トリピークスWebアウトプット
type TriPeaksWebOutput struct {
	Layout       [][]*TriPeaksWebOutputCard `json:"layout"`
	StockCount   int                        `json:"stockCount"`
	Waste        []*WebOutputCard           `json:"waste"`
	Phase        int                        `json:"phase"`
	MoveCount    int                        `json:"moveCount"`
	CanUndo      bool                       `json:"canUndo"`
	IsStalemate  bool                       `json:"isStalemate"`
	UndoToEscape int                        `json:"undoToEscape"`
	Hint         *TriPeaksWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
}

// TriPeaksWebController トリピークスWebコントローラークラス
type TriPeaksWebController = GameWebController[usecase.TriPeaksInteractorIF, TriPeaksWebInput, *TriPeaksWebOutput]

// NewTriPeaksWebController and NewTriPeaksWebControllerWithProvider are
// the standard and provider-backed constructors for TriPeaksWebController.
var NewTriPeaksWebController, NewTriPeaksWebControllerWithProvider = WebControllerPair[usecase.TriPeaksInteractorIF, TriPeaksWebInput, *TriPeaksWebOutput](
	newTriPeaksDefaultOutput, triPeaksDispatch,
)

func newTriPeaksDefaultOutput(msg string) *TriPeaksWebOutput {
	return &TriPeaksWebOutput{
		Layout:        make([][]*TriPeaksWebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func triPeaksDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TriPeaksInteractorIF, param TriPeaksWebInput, newDefault func(string) *TriPeaksWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.Reset())
	case "d", "draw":
		bc.writePresenterResponse(w, ti.Draw())
	case "rm", "remove":
		if param.Row == nil || param.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: row and col are required."))
			return true
		}
		bc.writePresenterResponse(w, ti.Remove(*param.Row, *param.Col))
	case "g", "giveup":
		bc.writePresenterResponse(w, ti.GiveUp())
	case "u", "undo":
		bc.writePresenterResponse(w, ti.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, ti.UndoN(*param.N))
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
