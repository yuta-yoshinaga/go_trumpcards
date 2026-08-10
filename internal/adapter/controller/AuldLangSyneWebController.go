//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AuldLangSyneWebInput オールド・ラング・サインWebインプット
type AuldLangSyneWebInput struct {
	BaseWebInput
	From *AuldLangSyneWebZone `json:"from,omitempty"`
	To   *AuldLangSyneWebZone `json:"to,omitempty"`
}

// AuldLangSyneWebZone ゾーン指定
type AuldLangSyneWebZone struct {
	Zone string `json:"zone"`
	Idx  *int   `json:"idx,omitempty"`
}

// AuldLangSyneWebOutputHint ヒント出力
type AuldLangSyneWebOutputHint struct {
	WasteIdx      int `json:"wasteIdx"`
	FoundationIdx int `json:"foundationIdx"`
}

// AuldLangSyneWebOutput オールド・ラング・サインWebアウトプット
type AuldLangSyneWebOutput struct {
	Foundations [][]*WebOutputCard         `json:"foundations"`
	Wastes      [][]*WebOutputCard         `json:"wastes"`
	StockCount  int                        `json:"stockCount"`
	Hint        *AuldLangSyneWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// AuldLangSyneWebController オールド・ラング・サインWebコントローラークラス
type AuldLangSyneWebController = GameWebController[usecase.AuldLangSyneInteractorIF, AuldLangSyneWebInput, *AuldLangSyneWebOutput]

// NewAuldLangSyneWebController and NewAuldLangSyneWebControllerWithProvider are
// the standard and provider-backed constructors for AuldLangSyneWebController.
var NewAuldLangSyneWebController, NewAuldLangSyneWebControllerWithProvider = webControllerPair[usecase.AuldLangSyneInteractorIF, AuldLangSyneWebInput, *AuldLangSyneWebOutput](
	newAuldLangSyneDefaultOutput, auldLangSyneDispatch,
)

func newAuldLangSyneDefaultOutput(msg string) *AuldLangSyneWebOutput {
	return &AuldLangSyneWebOutput{
		Foundations:   make([][]*WebOutputCard, 0),
		Wastes:        make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func auldLangSyneDispatch(bc *baseController, w http.ResponseWriter, ci usecase.AuldLangSyneInteractorIF, param AuldLangSyneWebInput, newDefault func(string) *AuldLangSyneWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return auldLangSyneMoveDispatch(bc, w, ci, param, newDefault)
	case "d", "deal":
		bc.writePresenterResponse(w, ci.Deal())
	case "g", "giveup":
		bc.writePresenterResponse(w, ci.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ci.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ci.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ci.Reset, ci.Hint, ci.ActionLog)
	}
	return true
}

// auldLangSyneMoveDispatch handles the game's only move. There is no
// stock->waste case: the deal is forced onto all four wastes at once, so the
// stock is reached through the `deal` command rather than through a move.
func auldLangSyneMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.AuldLangSyneInteractorIF, param AuldLangSyneWebInput, newDefault func(string) *AuldLangSyneWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	if param.From.Zone != "waste" || param.To.Zone != "foundation" {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
		return true
	}
	if !requireParam(bc, w, newDefault, param.From.Idx == nil || param.To.Idx == nil, "param error: from.idx and to.idx are required.") {
		return true
	}
	bc.writePresenterResponse(w, ci.PlayWasteToFoundation(*param.From.Idx, *param.To.Idx))
	return true
}
