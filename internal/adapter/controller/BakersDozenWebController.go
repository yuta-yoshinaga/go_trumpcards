//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BakersDozenWebInput ベーカーズダズンWebインプット
type BakersDozenWebInput struct {
	BaseWebInput
	From *BakersDozenWebZone `json:"from,omitempty"`
	To   *BakersDozenWebZone `json:"to,omitempty"`
}

// BakersDozenWebZone ゾーン指定
type BakersDozenWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// BakersDozenWebOutputTableauCard タブローカード出力
type BakersDozenWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// BakersDozenWebOutputHint ヒント出力
type BakersDozenWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// BakersDozenWebOutput ベーカーズダズンWebアウトプット
type BakersDozenWebOutput struct {
	Tableau    [][]*BakersDozenWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                   `json:"foundation"`
	Hint       *BakersDozenWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// BakersDozenWebController ベーカーズダズンWebコントローラークラス
type BakersDozenWebController = GameWebController[usecase.BakersDozenInteractorIF, BakersDozenWebInput, *BakersDozenWebOutput]

// NewBakersDozenWebController and NewBakersDozenWebControllerWithProvider are
// the standard and provider-backed constructors for BakersDozenWebController.
var NewBakersDozenWebController, NewBakersDozenWebControllerWithProvider = webControllerPair[usecase.BakersDozenInteractorIF, BakersDozenWebInput, *BakersDozenWebOutput](
	newBakersDozenDefaultOutput, bakersDozenDispatch,
)

func newBakersDozenDefaultOutput(msg string) *BakersDozenWebOutput {
	return &BakersDozenWebOutput{
		Tableau:       make([][]*BakersDozenWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func bakersDozenDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BakersDozenInteractorIF, param BakersDozenWebInput, newDefault func(string) *BakersDozenWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return bakersDozenMoveDispatch(bc, w, bi, param, newDefault)
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

func bakersDozenMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BakersDozenInteractorIF, param BakersDozenWebInput, newDefault func(string) *BakersDozenWebOutput) bool {
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
