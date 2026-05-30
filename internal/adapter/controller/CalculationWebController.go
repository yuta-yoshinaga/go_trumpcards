package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CalculationWebInput カルキュレーションWebインプット
type CalculationWebInput struct {
	BaseWebInput
	From *CalculationWebZone `json:"from,omitempty"`
	To   *CalculationWebZone `json:"to,omitempty"`
}

// CalculationWebZone ゾーン指定
type CalculationWebZone struct {
	Zone string `json:"zone"`
	Idx  *int   `json:"idx,omitempty"`
}

// CalculationWebOutputHint ヒント出力
type CalculationWebOutputHint struct {
	FromZone      string `json:"fromZone"`
	WasteIdx      int    `json:"wasteIdx"`
	FoundationIdx int    `json:"foundationIdx"`
}

// CalculationWebOutput カルキュレーションWebアウトプット
type CalculationWebOutput struct {
	Foundations [][]*WebOutputCard        `json:"foundations"`
	Wastes      [][]*WebOutputCard        `json:"wastes"`
	StockCount  int                       `json:"stockCount"`
	StockTop    *WebOutputCard            `json:"stockTop,omitempty"`
	Hint        *CalculationWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// CalculationWebController カルキュレーションWebコントローラークラス
type CalculationWebController = GameWebController[usecase.CalculationInteractorIF, CalculationWebInput, *CalculationWebOutput]

// NewCalculationWebController and NewCalculationWebControllerWithProvider are
// the standard and provider-backed constructors for CalculationWebController.
var NewCalculationWebController, NewCalculationWebControllerWithProvider = webControllerPair[usecase.CalculationInteractorIF, CalculationWebInput, *CalculationWebOutput](
	newCalculationDefaultOutput, calculationDispatch,
)

func newCalculationDefaultOutput(msg string) *CalculationWebOutput {
	return &CalculationWebOutput{
		Foundations:   make([][]*WebOutputCard, 0),
		Wastes:        make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func calculationDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CalculationInteractorIF, param CalculationWebInput, newDefault func(string) *CalculationWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return calculationMoveDispatch(bc, w, ci, param, newDefault)
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

func calculationMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CalculationInteractorIF, param CalculationWebInput, newDefault func(string) *CalculationWebOutput) bool {
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
