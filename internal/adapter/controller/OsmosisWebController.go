//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OsmosisWebInput オズモシスWebインプット
type OsmosisWebInput struct {
	BaseWebInput
	From *OsmosisWebZone `json:"from,omitempty"`
	To   *OsmosisWebZone `json:"to,omitempty"`
}

// OsmosisWebZone ゾーン指定
type OsmosisWebZone struct {
	Zone string `json:"zone"`
	Col  *int   `json:"col,omitempty"`
}

// OsmosisWebOutputHint ヒント出力
type OsmosisWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromCol  int    `json:"fromCol"`
	ToCol    int    `json:"toCol"`
}

// OsmosisWebOutput オズモシスWebアウトプット
type OsmosisWebOutput struct {
	Reserve    [][]*WebOutputCard `json:"reserve"`
	StockCount int                `json:"stockCount"`
	Waste      []*WebOutputCard   `json:"waste"`
	Foundation [][]*WebOutputCard `json:"foundation"`
	BaseRank   int                `json:"baseRank"`
	Phase      int                `json:"phase"`
	MoveCount  int                `json:"moveCount"`
	CanUndo    bool               `json:"canUndo"`
	// IsStalemate はどのカードもファンデーションへ送れなくなった状態 (#4808)。
	IsStalemate bool                  `json:"isStalemate"`
	Hint        *OsmosisWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// OsmosisWebController オズモシスWebコントローラークラス
type OsmosisWebController = GameWebController[usecase.OsmosisInteractorIF, OsmosisWebInput, *OsmosisWebOutput]

// NewOsmosisWebController and NewOsmosisWebControllerWithProvider are the
// standard and provider-backed constructors for OsmosisWebController.
var NewOsmosisWebController, NewOsmosisWebControllerWithProvider = webControllerPair[usecase.OsmosisInteractorIF, OsmosisWebInput, *OsmosisWebOutput](
	newOsmosisDefaultOutput, osmosisDispatch,
)

func newOsmosisDefaultOutput(msg string) *OsmosisWebOutput {
	return &OsmosisWebOutput{
		Reserve:       make([][]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func osmosisDispatch(bc *baseController, w http.ResponseWriter, oi usecase.OsmosisInteractorIF, param OsmosisWebInput, newDefault func(string) *OsmosisWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, oi.Draw())
	case "m", "move":
		return osmosisMoveDispatch(bc, w, oi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, oi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, oi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, oi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, oi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, oi.Reset, oi.Hint, oi.ActionLog)
	}
	return true
}

func osmosisMoveDispatch(bc *baseController, w http.ResponseWriter, oi usecase.OsmosisInteractorIF, param OsmosisWebInput, newDefault func(string) *OsmosisWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "waste" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, oi.MoveWasteToFoundation(*param.To.Col))
	case fromZone == "reserve" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, oi.MoveReserveToFoundation(*param.From.Col, *param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
