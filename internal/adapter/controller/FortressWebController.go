//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FortressWebInput Fortress Web インプット
type FortressWebInput struct {
	BaseWebInput
	From *FortressWebZone `json:"from,omitempty"`
	To   *FortressWebZone `json:"to,omitempty"`
}

// FortressWebZone ゾーン指定
type FortressWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// FortressWebOutputTableauCard タブローカード出力
type FortressWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// FortressWebOutputHint ヒント出力
type FortressWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// FortressWebOutput Fortress Web アウトプット
type FortressWebOutput struct {
	Tableau    [][]*FortressWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                `json:"foundation"`
	Hint       *FortressWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// FortressWebController Fortress Web コントローラークラス
type FortressWebController = GameWebController[usecase.FortressInteractorIF, FortressWebInput, *FortressWebOutput]

// NewFortressWebController and NewFortressWebControllerWithProvider are
// the standard and provider-backed constructors for FortressWebController.
var NewFortressWebController, NewFortressWebControllerWithProvider = webControllerPair[usecase.FortressInteractorIF, FortressWebInput, *FortressWebOutput](
	newFortressDefaultOutput, fortressDispatch,
)

func newFortressDefaultOutput(msg string) *FortressWebOutput {
	return &FortressWebOutput{
		Tableau:       make([][]*FortressWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func fortressDispatch(bc *baseController, w http.ResponseWriter, bi usecase.FortressInteractorIF, param FortressWebInput, newDefault func(string) *FortressWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return fortressMoveDispatch(bc, w, bi, param, newDefault)
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

func fortressMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.FortressInteractorIF, param FortressWebInput, newDefault func(string) *FortressWebOutput) bool {
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
