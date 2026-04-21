package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GolfWebInput ゴルフソリティアWebインプット
type GolfWebInput struct {
	BaseWebInput
	Col *int `json:"col,omitempty"`
}

// GolfWebOutputCard タブローカード出力
type GolfWebOutputCard struct {
	Card    *WebOutputCard `json:"card"`
	Removed bool           `json:"removed"`
	Exposed bool           `json:"exposed"`
}

// GolfWebOutputHint ヒント出力
type GolfWebOutputHint struct {
	Type string `json:"type"`
	Col  int    `json:"col"`
}

// GolfWebOutput ゴルフソリティアWebアウトプット
type GolfWebOutput struct {
	Layout       [][]*GolfWebOutputCard `json:"layout"`
	StockCount   int                    `json:"stockCount"`
	Waste        []*WebOutputCard       `json:"waste"`
	Phase        int                    `json:"phase"`
	MoveCount    int                    `json:"moveCount"`
	CanUndo      bool                   `json:"canUndo"`
	IsStalemate  bool                   `json:"isStalemate"`
	UndoToEscape int                    `json:"undoToEscape"`
	Hint         *GolfWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
}

// GolfWebController ゴルフソリティアWebコントローラークラス
type GolfWebController = GameWebController[usecase.GolfInteractorIF, GolfWebInput, *GolfWebOutput]

// NewGolfWebController and NewGolfWebControllerWithProvider are
// the standard and provider-backed constructors for GolfWebController.
var NewGolfWebController, NewGolfWebControllerWithProvider = webControllerPair[usecase.GolfInteractorIF, GolfWebInput, *GolfWebOutput](
	newGolfDefaultOutput, golfDispatch,
)

func newGolfDefaultOutput(msg string) *GolfWebOutput {
	return &GolfWebOutput{
		Layout:        make([][]*GolfWebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func golfDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GolfInteractorIF, param GolfWebInput, newDefault func(string) *GolfWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, gi.Draw())
	case "rm", "remove":
		if param.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: col is required."))
			return true
		}
		bc.writePresenterResponse(w, gi.Remove(*param.Col))
	case "g", "giveup":
		bc.writePresenterResponse(w, gi.GiveUp())
	case "u", "undo":
		bc.writePresenterResponse(w, gi.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, gi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, gi.Reset, gi.Hint, gi.ActionLog)
	}
	return true
}
