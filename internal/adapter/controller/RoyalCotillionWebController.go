//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RoyalCotillionWebInput ロイヤルコティヨン Web インプット
type RoyalCotillionWebInput struct {
	BaseWebInput
	From *RoyalCotillionWebZone `json:"from,omitempty"`
	To   *RoyalCotillionWebZone `json:"to,omitempty"`
}

// RoyalCotillionWebZone ゾーン指定。Zone は "tableau" / "reserve" / "waste" / "stock" / "foundation"。
type RoyalCotillionWebZone struct {
	Zone string `json:"zone"`
	// Col はタブロー山（0..7）。捨て札・山札・基礎札では不要。
	Col *int `json:"col,omitempty"`
}

// RoyalCotillionWebOutputHint ヒント出力
type RoyalCotillionWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// RoyalCotillionWebOutput ロイヤルコティヨン Web アウトプット
type RoyalCotillionWebOutput struct {
	// Tableau は 1 枠 1 枚。空き枠は null。
	Tableau []*WebOutputCard `json:"tableau"`
	// Reserve は 3 枚重ねの山が 4 つ。一番上だけが使え、空いたら二度と埋まらない。
	Reserve [][]*WebOutputCard `json:"reserve"`
	// FoundationOdd は基礎札ごとの系統。true が A 始まり、false が 2 始まり。
	FoundationOdd []bool                       `json:"foundationOdd"`
	Foundation    [][]*WebOutputCard           `json:"foundation"`
	StockCount    int                          `json:"stockCount"`
	Waste         []*WebOutputCard             `json:"waste"`
	Hint          *RoyalCotillionWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// RoyalCotillionWebController ロイヤルコティヨン Web コントローラークラス
type RoyalCotillionWebController = GameWebController[usecase.RoyalCotillionInteractorIF, RoyalCotillionWebInput, *RoyalCotillionWebOutput]

// NewRoyalCotillionWebController and NewRoyalCotillionWebControllerWithProvider are the
// standard and provider-backed constructors for RoyalCotillionWebController.
var NewRoyalCotillionWebController, NewRoyalCotillionWebControllerWithProvider = webControllerPair[usecase.RoyalCotillionInteractorIF, RoyalCotillionWebInput, *RoyalCotillionWebOutput](
	newRoyalCotillionDefaultOutput, royalcotillionDispatch,
)

func newRoyalCotillionDefaultOutput(msg string) *RoyalCotillionWebOutput {
	return &RoyalCotillionWebOutput{
		Tableau:       make([]*WebOutputCard, 0),
		Reserve:       make([][]*WebOutputCard, 0),
		FoundationOdd: make([]bool, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func royalcotillionDispatch(bc *baseController, w http.ResponseWriter, ci usecase.RoyalCotillionInteractorIF, param RoyalCotillionWebInput, newDefault func(string) *RoyalCotillionWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "m", "move":
		return royalcotillionMoveDispatch(bc, w, ci, param, newDefault)
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

func royalcotillionMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.RoyalCotillionInteractorIF, param RoyalCotillionWebInput, newDefault func(string) *RoyalCotillionWebOutput) bool {
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
	case fromZone == "reserve" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveReserveToFoundation(*param.From.Col))
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
