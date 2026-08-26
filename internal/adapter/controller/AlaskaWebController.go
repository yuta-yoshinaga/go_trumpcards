//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AlaskaWebInput アラスカWebインプット
type AlaskaWebInput struct {
	BaseWebInput
	From *AlaskaWebZone `json:"from,omitempty"`
	To   *AlaskaWebZone `json:"to,omitempty"`
}

// AlaskaWebZone ゾーン指定
type AlaskaWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// AlaskaWebOutputHint ヒント出力
type AlaskaWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// AlaskaWebOutputTableauCard タブローカード出力。Alaska は Klondike とは別バケット
// (extra4 worker) なので AlaskaWebOutputTableauCard を共有せず独自に定義する。
// JSON のフィールド名は Klondike と同一なので、フロントの型は変わらない。
type AlaskaWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// AlaskaWebOutput アラスカWebアウトプット
type AlaskaWebOutput struct {
	Tableau    [][]*AlaskaWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard              `json:"foundation"`
	Hint       *AlaskaWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// AlaskaWebController アラスカWebコントローラークラス
type AlaskaWebController = GameWebController[usecase.AlaskaInteractorIF, AlaskaWebInput, *AlaskaWebOutput]

// NewAlaskaWebController and NewAlaskaWebControllerWithProvider are
// the standard and provider-backed constructors for AlaskaWebController.
var NewAlaskaWebController, NewAlaskaWebControllerWithProvider = webControllerPair[usecase.AlaskaInteractorIF, AlaskaWebInput, *AlaskaWebOutput](
	newAlaskaDefaultOutput, alaskaDispatch,
)

func newAlaskaDefaultOutput(msg string) *AlaskaWebOutput {
	return &AlaskaWebOutput{
		Tableau:       make([][]*AlaskaWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func alaskaDispatch(bc *baseController, w http.ResponseWriter, ri usecase.AlaskaInteractorIF, param AlaskaWebInput, newDefault func(string) *AlaskaWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return alaskaMoveDispatch(bc, w, ri, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ri.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ri.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ri.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, ri.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ri.Reset, ri.Hint, ri.ActionLog)
	}
	return true
}

func alaskaMoveDispatch(bc *baseController, w http.ResponseWriter, ri usecase.AlaskaInteractorIF, param AlaskaWebInput, newDefault func(string) *AlaskaWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil, "param error: from.col, from.cardIndex, to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, ri.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ri.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
