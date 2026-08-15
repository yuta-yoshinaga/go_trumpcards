//go:build !js || !wasm || solo

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
	Hint           *ScorpionWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
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
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, si.Reset, si.Hint, si.ActionLog)
	}
	return true
}

func scorpionMoveDispatch(bc *baseController, w http.ResponseWriter, si usecase.ScorpionInteractorIF, param ScorpionWebInput, newDefault func(string) *ScorpionWebOutput) bool {
	mv := tableauMove{haveFrom: param.From != nil, haveTo: param.To != nil}
	if param.From != nil {
		mv.fromZone, mv.fromCol, mv.fromCardIndex = param.From.Zone, param.From.Col, param.From.CardIndex
	}
	if param.To != nil {
		mv.toZone, mv.toCol = param.To.Zone, param.To.Col
	}
	return dispatchTableauOnlyMove(bc, w, mv, si.MoveTableauToTableau, newDefault)
}
