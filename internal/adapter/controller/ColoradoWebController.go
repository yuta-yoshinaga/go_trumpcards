//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ColoradoWebInput コロラド Web インプット
type ColoradoWebInput struct {
	BaseWebInput
	From *ColoradoWebZone `json:"from,omitempty"`
	To   *ColoradoWebZone `json:"to,omitempty"`
}

// ColoradoWebZone ゾーン指定。Zone は "tableau" / "waste" / "stock" / "foundation"。
type ColoradoWebZone struct {
	Zone string `json:"zone"`
	// Col はタブロー山（0..19）。捨て札・山札・基礎札では不要。
	Col *int `json:"col,omitempty"`
}

// ColoradoWebOutputHint ヒント出力
type ColoradoWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// ColoradoWebOutput コロラド Web アウトプット
type ColoradoWebOutput struct {
	Tableau    [][]*WebOutputCard `json:"tableau"`
	Foundation [][]*WebOutputCard `json:"foundation"`
	// FoundationAscending は基礎札ごとの積む向き。true が A→K、false が K→A。
	// 添字から推測させると、並びを変えたときに表示だけが静かにずれる。
	FoundationAscending []bool                 `json:"foundationAscending"`
	StockCount          int                    `json:"stockCount"`
	Waste               []*WebOutputCard       `json:"waste"`
	Hint                *ColoradoWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// ColoradoWebController コロラド Web コントローラークラス
type ColoradoWebController = GameWebController[usecase.ColoradoInteractorIF, ColoradoWebInput, *ColoradoWebOutput]

// NewColoradoWebController and NewColoradoWebControllerWithProvider are the
// standard and provider-backed constructors for ColoradoWebController.
var NewColoradoWebController, NewColoradoWebControllerWithProvider = webControllerPair[usecase.ColoradoInteractorIF, ColoradoWebInput, *ColoradoWebOutput](
	newColoradoDefaultOutput, coloradoDispatch,
)

func newColoradoDefaultOutput(msg string) *ColoradoWebOutput {
	return &ColoradoWebOutput{
		Tableau:             make([][]*WebOutputCard, 0),
		Foundation:          make([][]*WebOutputCard, 0),
		FoundationAscending: make([]bool, 0),
		Waste:               make([]*WebOutputCard, 0),
		WebOutputBase:       WebOutputBase{Message: msg},
	}
}

func coloradoDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ColoradoInteractorIF, param ColoradoWebInput, newDefault func(string) *ColoradoWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "m", "move":
		return coloradoMoveDispatch(bc, w, ci, param, newDefault)
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

func coloradoMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ColoradoInteractorIF, param ColoradoWebInput, newDefault func(string) *ColoradoWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, ci.MoveWasteToFoundation())
	case fromZone == "waste" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveWasteToTableau(*param.To.Col))
	// The stock can fill a gap without going through the waste, which matters
	// because the single pass makes every turned card expensive.
	case fromZone == "stock" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveStockToTableau(*param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
