//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SomersetWebInput Somerset Web インプット
type SomersetWebInput struct {
	BaseWebInput
	From *SomersetWebZone `json:"from,omitempty"`
	To   *SomersetWebZone `json:"to,omitempty"`
}

// SomersetWebZone ゾーン指定
type SomersetWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// SomersetWebOutputTableauCard タブローカード出力
type SomersetWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// SomersetWebOutputHint ヒント出力
type SomersetWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// SomersetWebOutput Somerset Web アウトプット
type SomersetWebOutput struct {
	Tableau    [][]*SomersetWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                `json:"foundation"`
	Hint       *SomersetWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// SomersetWebController Somerset Web コントローラークラス
type SomersetWebController = GameWebController[usecase.SomersetInteractorIF, SomersetWebInput, *SomersetWebOutput]

// NewSomersetWebController and NewSomersetWebControllerWithProvider are
// the standard and provider-backed constructors for SomersetWebController.
var NewSomersetWebController, NewSomersetWebControllerWithProvider = webControllerPair[usecase.SomersetInteractorIF, SomersetWebInput, *SomersetWebOutput](
	newSomersetDefaultOutput, somersetDispatch,
)

func newSomersetDefaultOutput(msg string) *SomersetWebOutput {
	return &SomersetWebOutput{
		Tableau:       make([][]*SomersetWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func somersetDispatch(bc *baseController, w http.ResponseWriter, bi usecase.SomersetInteractorIF, param SomersetWebInput, newDefault func(string) *SomersetWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return somersetMoveDispatch(bc, w, bi, param, newDefault)
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

func somersetMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.SomersetInteractorIF, param SomersetWebInput, newDefault func(string) *SomersetWebOutput) bool {
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
