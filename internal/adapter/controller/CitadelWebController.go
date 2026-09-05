//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CitadelWebInput Citadel Web インプット
type CitadelWebInput struct {
	BaseWebInput
	From *CitadelWebZone `json:"from,omitempty"`
	To   *CitadelWebZone `json:"to,omitempty"`
}

// CitadelWebZone ゾーン指定
type CitadelWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// CitadelWebOutputTableauCard タブローカード出力
type CitadelWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// CitadelWebOutputHint ヒント出力
type CitadelWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// CitadelWebOutput Citadel Web アウトプット
type CitadelWebOutput struct {
	Tableau    [][]*CitadelWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard               `json:"foundation"`
	Hint       *CitadelWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// CitadelWebController Citadel Web コントローラークラス
type CitadelWebController = GameWebController[usecase.CitadelInteractorIF, CitadelWebInput, *CitadelWebOutput]

// NewCitadelWebController and NewCitadelWebControllerWithProvider are
// the standard and provider-backed constructors for CitadelWebController.
var NewCitadelWebController, NewCitadelWebControllerWithProvider = webControllerPair[usecase.CitadelInteractorIF, CitadelWebInput, *CitadelWebOutput](
	newCitadelDefaultOutput, citadelDispatch,
)

func newCitadelDefaultOutput(msg string) *CitadelWebOutput {
	return &CitadelWebOutput{
		Tableau:       make([][]*CitadelWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func citadelDispatch(bc *baseController, w http.ResponseWriter, bi usecase.CitadelInteractorIF, param CitadelWebInput, newDefault func(string) *CitadelWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return citadelMoveDispatch(bc, w, bi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, bi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, bi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, bi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, bi.Reset, bi.Hint, bi.ActionLog)
	}
	return true
}

func citadelMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.CitadelInteractorIF, param CitadelWebInput, newDefault func(string) *CitadelWebOutput) bool {
	mv := topCardMove{haveFrom: param.From != nil, haveTo: param.To != nil}
	if param.From != nil {
		mv.fromZone, mv.fromCol = param.From.Zone, param.From.Col
	}
	if param.To != nil {
		mv.toZone, mv.toCol = param.To.Zone, param.To.Col
	}
	return dispatchTopCardMove(bc, w, mv, topCardMoveFns{
		tableauToTableau:    bi.MoveTableauToTableau,
		tableauToFoundation: bi.MoveTableauToFoundation,
	}, newDefault)
}
