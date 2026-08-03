//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CongressWebInput コングレス Web インプット
type CongressWebInput struct {
	BaseWebInput
	From *CongressWebZone `json:"from,omitempty"`
	To   *CongressWebZone `json:"to,omitempty"`
}

// CongressWebZone ゾーン指定。Zone は "tableau" / "waste" / "stock" / "foundation"。
type CongressWebZone struct {
	Zone string `json:"zone"`
	// Col はタブロー山（0..7）。捨て札・山札・基礎札では不要。
	Col *int `json:"col,omitempty"`
}

// CongressWebOutputHint ヒント出力
type CongressWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// CongressWebOutput コングレス Web アウトプット
type CongressWebOutput struct {
	Tableau    [][]*WebOutputCard     `json:"tableau"`
	Foundation [][]*WebOutputCard     `json:"foundation"`
	StockCount int                    `json:"stockCount"`
	Waste      []*WebOutputCard       `json:"waste"`
	Hint       *CongressWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// CongressWebController コングレス Web コントローラークラス
type CongressWebController = GameWebController[usecase.CongressInteractorIF, CongressWebInput, *CongressWebOutput]

// NewCongressWebController and NewCongressWebControllerWithProvider are the
// standard and provider-backed constructors for CongressWebController.
var NewCongressWebController, NewCongressWebControllerWithProvider = webControllerPair[usecase.CongressInteractorIF, CongressWebInput, *CongressWebOutput](
	newCongressDefaultOutput, congressDispatch,
)

func newCongressDefaultOutput(msg string) *CongressWebOutput {
	return &CongressWebOutput{
		Tableau:       make([][]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func congressDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CongressInteractorIF, param CongressWebInput, newDefault func(string) *CongressWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "m", "move":
		return congressMoveDispatch(bc, w, ci, param, newDefault)
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

func congressMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CongressInteractorIF, param CongressWebInput, newDefault func(string) *CongressWebOutput) bool {
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
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToTableau(*param.From.Col, *param.To.Col))
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
