//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MatrimonyWebInput マトリモニー Web インプット
type MatrimonyWebInput struct {
	BaseWebInput
	From *MatrimonyWebZone `json:"from,omitempty"`
	To   *MatrimonyWebZone `json:"to,omitempty"`
}

// MatrimonyWebZone identifies a game zone.
type MatrimonyWebZone struct {
	Zone string `json:"zone"`
	// Col はタブロー枠（0..15）。捨て札・山札・基礎札では不要。
	Col *int `json:"col,omitempty"`
}

// MatrimonyWebOutputHint ヒント出力
type MatrimonyWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// MatrimonyWebOutput マトリモニー Web アウトプット
type MatrimonyWebOutput struct {
	// Tableau は 1 枠 1 枚。空き枠は null。
	Tableau     []*WebOutputCard        `json:"tableau"`
	Foundation  [][]*WebOutputCard      `json:"foundation"`
	StockCount  int                     `json:"stockCount"`
	RedealCount int                     `json:"redealCount"`
	Waste       []*WebOutputCard        `json:"waste"`
	Hint        *MatrimonyWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// MatrimonyWebController マトリモニー Web コントローラークラス
type MatrimonyWebController = GameWebController[usecase.MatrimonyInteractorIF, MatrimonyWebInput, *MatrimonyWebOutput]

// NewMatrimonyWebController and NewMatrimonyWebControllerWithProvider are the
// standard and provider-backed constructors for MatrimonyWebController.
var NewMatrimonyWebController, NewMatrimonyWebControllerWithProvider = webControllerPair[usecase.MatrimonyInteractorIF, MatrimonyWebInput, *MatrimonyWebOutput](
	newMatrimonyDefaultOutput, matrimonyDispatch,
)

func newMatrimonyDefaultOutput(msg string) *MatrimonyWebOutput {
	return &MatrimonyWebOutput{
		Tableau:       make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func matrimonyDispatch(bc *baseController, w http.ResponseWriter, ci usecase.MatrimonyInteractorIF, param MatrimonyWebInput, newDefault func(string) *MatrimonyWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "m", "move":
		return matrimonyMoveDispatch(bc, w, ci, param, newDefault)
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

func matrimonyMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.MatrimonyInteractorIF, param MatrimonyWebInput, newDefault func(string) *MatrimonyWebOutput) bool {
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
