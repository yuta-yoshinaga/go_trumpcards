//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EasthavenWebInput イーストヘイブンWebインプット
type EasthavenWebInput struct {
	BaseWebInput
	From *EasthavenWebZone `json:"from,omitempty"`
	To   *EasthavenWebZone `json:"to,omitempty"`
}

// EasthavenWebZone ゾーン指定
type EasthavenWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// EasthavenWebOutputHint ヒント出力
type EasthavenWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// EasthavenWebOutput イーストヘイブンWebアウトプット
type EasthavenWebOutput struct {
	Tableau    [][]*KlondikeWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                `json:"foundation"`
	StockCount int                               `json:"stockCount"`
	Hint       *EasthavenWebOutputHint           `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// EasthavenWebController イーストヘイブンWebコントローラークラス
type EasthavenWebController = GameWebController[usecase.EasthavenInteractorIF, EasthavenWebInput, *EasthavenWebOutput]

// NewEasthavenWebController and NewEasthavenWebControllerWithProvider are
// the standard and provider-backed constructors for EasthavenWebController.
var NewEasthavenWebController, NewEasthavenWebControllerWithProvider = webControllerPair[usecase.EasthavenInteractorIF, EasthavenWebInput, *EasthavenWebOutput](
	newEasthavenDefaultOutput, easthavenDispatch,
)

func newEasthavenDefaultOutput(msg string) *EasthavenWebOutput {
	return &EasthavenWebOutput{
		Tableau:       make([][]*KlondikeWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func easthavenDispatch(bc *baseController, w http.ResponseWriter, ei usecase.EasthavenInteractorIF, param EasthavenWebInput, newDefault func(string) *EasthavenWebOutput) bool {
	switch param.Command {
	case "d", "deal":
		bc.writePresenterResponse(w, ei.Deal())
	case "m", "move":
		return easthavenMoveDispatch(bc, w, ei, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ei.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ei.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ei.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, ei.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ei.Reset, ei.Hint, ei.ActionLog)
	}
	return true
}

func easthavenMoveDispatch(bc *baseController, w http.ResponseWriter, ei usecase.EasthavenInteractorIF, param EasthavenWebInput, newDefault func(string) *EasthavenWebOutput) bool {
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
		bc.writePresenterResponse(w, ei.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ei.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
