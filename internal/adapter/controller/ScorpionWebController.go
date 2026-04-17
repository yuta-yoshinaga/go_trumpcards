package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ScorpionWebInput スコーピオンWebインプット
type ScorpionWebInput struct {
	BaseWebInput
	From *ScorpionWebZone `json:"from,omitempty"`
	To   *ScorpionWebZone `json:"to,omitempty"`
}

// ScorpionWebZone ゾーン指定
type ScorpionWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// ScorpionWebOutputHint ヒント出力
type ScorpionWebOutputHint struct {
	FromCol   int `json:"fromCol"`
	CardIndex int `json:"cardIndex"`
	ToCol     int `json:"toCol"`
}

// ScorpionWebOutput スコーピオンWebアウトプット
type ScorpionWebOutput struct {
	Tableau        [][]*KlondikeWebOutputTableauCard `json:"tableau"`
	StockCount     int                               `json:"stockCount"`
	CompletedSuits int                               `json:"completedSuits"`
	Phase          int                               `json:"phase"`
	MoveCount      int                               `json:"moveCount"`
	CanUndo        bool                              `json:"canUndo"`
	IsStalemate    bool                              `json:"isStalemate"`
	UndoToEscape   int                               `json:"undoToEscape"`
	Hint           *ScorpionWebOutputHint            `json:"hint,omitempty"`
	WebOutputBase
}

// ScorpionWebController スコーピオンWebコントローラークラス
type ScorpionWebController = GameWebController[usecase.ScorpionInteractorIF, ScorpionWebInput, *ScorpionWebOutput]

// NewScorpionWebController and NewScorpionWebControllerWithProvider are
// the standard and provider-backed constructors for ScorpionWebController.
var NewScorpionWebController, NewScorpionWebControllerWithProvider = webControllerPair[usecase.ScorpionInteractorIF, ScorpionWebInput, *ScorpionWebOutput](
	newScorpionDefaultOutput, scorpionDispatch,
)

func newScorpionDefaultOutput(msg string) *ScorpionWebOutput {
	return &ScorpionWebOutput{
		Tableau:       make([][]*KlondikeWebOutputTableauCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func scorpionDispatch(bc *baseController, w http.ResponseWriter, si usecase.ScorpionInteractorIF, param ScorpionWebInput, newDefault func(string) *ScorpionWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.Reset())
	case "d", "deal":
		bc.writePresenterResponse(w, si.Deal())
	case "m", "move":
		return scorpionMoveDispatch(bc, w, si, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, si.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, si.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, si.UndoN(*param.N))
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}

func scorpionMoveDispatch(bc *baseController, w http.ResponseWriter, si usecase.ScorpionInteractorIF, param ScorpionWebInput, newDefault func(string) *ScorpionWebOutput) bool {
	if param.From == nil || param.To == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from and to are required."))
		return true
	}
	if param.From.Zone != "tableau" || param.To.Zone != "tableau" {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones. Only tableau to tableau is supported."))
		return true
	}
	if param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col, from.cardIndex, to.col are required."))
		return true
	}
	bc.writePresenterResponse(w, si.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	return true
}
