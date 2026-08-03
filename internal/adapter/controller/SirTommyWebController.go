//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SirTommyWebInput サー・トミーWebインプット
type SirTommyWebInput struct {
	BaseWebInput
	From *SirTommyWebZone `json:"from,omitempty"`
	To   *SirTommyWebZone `json:"to,omitempty"`
}

// SirTommyWebZone ゾーン指定
type SirTommyWebZone struct {
	Zone string `json:"zone"`
	Idx  *int   `json:"idx,omitempty"`
}

// SirTommyWebOutputHint ヒント出力
type SirTommyWebOutputHint struct {
	FromZone      string `json:"fromZone"`
	WasteIdx      int    `json:"wasteIdx"`
	FoundationIdx int    `json:"foundationIdx"`
}

// SirTommyWebOutput サー・トミーWebアウトプット
type SirTommyWebOutput struct {
	Foundations [][]*WebOutputCard     `json:"foundations"`
	Wastes      [][]*WebOutputCard     `json:"wastes"`
	StockCount  int                    `json:"stockCount"`
	StockTop    *WebOutputCard         `json:"stockTop,omitempty"`
	Hint        *SirTommyWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// SirTommyWebController サー・トミーWebコントローラークラス
type SirTommyWebController = GameWebController[usecase.SirTommyInteractorIF, SirTommyWebInput, *SirTommyWebOutput]

// NewSirTommyWebController and NewSirTommyWebControllerWithProvider are
// the standard and provider-backed constructors for SirTommyWebController.
var NewSirTommyWebController, NewSirTommyWebControllerWithProvider = webControllerPair[usecase.SirTommyInteractorIF, SirTommyWebInput, *SirTommyWebOutput](
	newSirTommyDefaultOutput, sirTommyDispatch,
)

func newSirTommyDefaultOutput(msg string) *SirTommyWebOutput {
	return &SirTommyWebOutput{
		Foundations:   make([][]*WebOutputCard, 0),
		Wastes:        make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func sirTommyDispatch(bc *baseController, w http.ResponseWriter, ci usecase.SirTommyInteractorIF, param SirTommyWebInput, newDefault func(string) *SirTommyWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return sirTommyMoveDispatch(bc, w, ci, param, newDefault)
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

func sirTommyMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.SirTommyInteractorIF, param SirTommyWebInput, newDefault func(string) *SirTommyWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "stock" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.To.Idx == nil, "param error: to.idx is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.PlayStockToFoundation(*param.To.Idx))
	case fromZone == "stock" && toZone == "waste":
		if !requireParam(bc, w, newDefault, param.To.Idx == nil, "param error: to.idx is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.PlayStockToWaste(*param.To.Idx))
	case fromZone == "waste" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Idx == nil || param.To.Idx == nil, "param error: from.idx and to.idx are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.PlayWasteToFoundation(*param.From.Idx, *param.To.Idx))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
