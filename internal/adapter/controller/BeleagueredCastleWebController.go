//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BeleagueredCastleWebInput Beleaguered Castle Web インプット
type BeleagueredCastleWebInput struct {
	BaseWebInput
	From *BeleagueredCastleWebZone `json:"from,omitempty"`
	To   *BeleagueredCastleWebZone `json:"to,omitempty"`
}

// BeleagueredCastleWebZone ゾーン指定
type BeleagueredCastleWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// BeleagueredCastleWebOutputTableauCard タブローカード出力
type BeleagueredCastleWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// BeleagueredCastleWebOutputHint ヒント出力
type BeleagueredCastleWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// BeleagueredCastleWebOutput Beleaguered Castle Web アウトプット
type BeleagueredCastleWebOutput struct {
	Tableau    [][]*BeleagueredCastleWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                         `json:"foundation"`
	Hint       *BeleagueredCastleWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// BeleagueredCastleWebController Beleaguered Castle Web コントローラークラス
type BeleagueredCastleWebController = GameWebController[usecase.BeleagueredCastleInteractorIF, BeleagueredCastleWebInput, *BeleagueredCastleWebOutput]

// NewBeleagueredCastleWebController and NewBeleagueredCastleWebControllerWithProvider are
// the standard and provider-backed constructors for BeleagueredCastleWebController.
var NewBeleagueredCastleWebController, NewBeleagueredCastleWebControllerWithProvider = webControllerPair[usecase.BeleagueredCastleInteractorIF, BeleagueredCastleWebInput, *BeleagueredCastleWebOutput](
	newBeleagueredCastleDefaultOutput, beleagueredCastleDispatch,
)

func newBeleagueredCastleDefaultOutput(msg string) *BeleagueredCastleWebOutput {
	return &BeleagueredCastleWebOutput{
		Tableau:       make([][]*BeleagueredCastleWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func beleagueredCastleDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BeleagueredCastleInteractorIF, param BeleagueredCastleWebInput, newDefault func(string) *BeleagueredCastleWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return beleagueredCastleMoveDispatch(bc, w, bi, param, newDefault)
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

func beleagueredCastleMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BeleagueredCastleInteractorIF, param BeleagueredCastleWebInput, newDefault func(string) *BeleagueredCastleWebOutput) bool {
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
