//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// StreetsAndAlleysWebInput Streets and Alleys Web インプット
type StreetsAndAlleysWebInput struct {
	BaseWebInput
	From *StreetsAndAlleysWebZone `json:"from,omitempty"`
	To   *StreetsAndAlleysWebZone `json:"to,omitempty"`
}

// StreetsAndAlleysWebZone ゾーン指定
type StreetsAndAlleysWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// StreetsAndAlleysWebOutputTableauCard タブローカード出力
type StreetsAndAlleysWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// StreetsAndAlleysWebOutputHint ヒント出力
type StreetsAndAlleysWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// StreetsAndAlleysWebOutput Streets and Alleys Web アウトプット
type StreetsAndAlleysWebOutput struct {
	Tableau    [][]*StreetsAndAlleysWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                        `json:"foundation"`
	Hint       *StreetsAndAlleysWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// StreetsAndAlleysWebController Streets and Alleys Web コントローラークラス
type StreetsAndAlleysWebController = GameWebController[usecase.StreetsAndAlleysInteractorIF, StreetsAndAlleysWebInput, *StreetsAndAlleysWebOutput]

// NewStreetsAndAlleysWebController and NewStreetsAndAlleysWebControllerWithProvider are
// the standard and provider-backed constructors for StreetsAndAlleysWebController.
var NewStreetsAndAlleysWebController, NewStreetsAndAlleysWebControllerWithProvider = webControllerPair[usecase.StreetsAndAlleysInteractorIF, StreetsAndAlleysWebInput, *StreetsAndAlleysWebOutput](
	newStreetsAndAlleysDefaultOutput, streetsAndAlleysDispatch,
)

func newStreetsAndAlleysDefaultOutput(msg string) *StreetsAndAlleysWebOutput {
	return &StreetsAndAlleysWebOutput{
		Tableau:       make([][]*StreetsAndAlleysWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func streetsAndAlleysDispatch(bc *baseController, w http.ResponseWriter, bi usecase.StreetsAndAlleysInteractorIF, param StreetsAndAlleysWebInput, newDefault func(string) *StreetsAndAlleysWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return streetsAndAlleysMoveDispatch(bc, w, bi, param, newDefault)
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

func streetsAndAlleysMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.StreetsAndAlleysInteractorIF, param StreetsAndAlleysWebInput, newDefault func(string) *StreetsAndAlleysWebOutput) bool {
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
