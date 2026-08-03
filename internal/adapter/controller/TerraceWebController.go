//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TerraceWebInput テラス Web インプット
type TerraceWebInput struct {
	BaseWebInput
	From *TerraceWebZone `json:"from,omitempty"`
	To   *TerraceWebZone `json:"to,omitempty"`
}

// TerraceWebZone ゾーン指定。Zone は "reserve" / "waste" / "tableau" / "foundation"。
type TerraceWebZone struct {
	Zone string `json:"zone"`
	// Col はタブロー山（0..8）。テラス・捨て札・基礎札では不要。
	Col *int `json:"col,omitempty"`
}

// TerraceWebOutputHint ヒント出力
type TerraceWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// TerraceWebOutput テラス Web アウトプット
type TerraceWebOutput struct {
	Reserve    []*WebOutputCard   `json:"reserve"`
	Tableau    [][]*WebOutputCard `json:"tableau"`
	Foundation [][]*WebOutputCard `json:"foundation"`
	StockCount int                `json:"stockCount"`
	Waste      []*WebOutputCard   `json:"waste"`
	// BaseRank は基礎札が始まるランク。0 は「まだ決まっていない」。
	BaseRank int `json:"baseRank"`
	// AwaitingBaseRank が真の間は、最初に基礎札へ送る 1 枚が開始ランクを決める。
	AwaitingBaseRank bool                  `json:"awaitingBaseRank"`
	Hint             *TerraceWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// TerraceWebController テラス Web コントローラークラス
type TerraceWebController = GameWebController[usecase.TerraceInteractorIF, TerraceWebInput, *TerraceWebOutput]

// NewTerraceWebController and NewTerraceWebControllerWithProvider are the
// standard and provider-backed constructors for TerraceWebController.
var NewTerraceWebController, NewTerraceWebControllerWithProvider = webControllerPair[usecase.TerraceInteractorIF, TerraceWebInput, *TerraceWebOutput](
	newTerraceDefaultOutput, terraceDispatch,
)

func newTerraceDefaultOutput(msg string) *TerraceWebOutput {
	return &TerraceWebOutput{
		Reserve:       make([]*WebOutputCard, 0),
		Tableau:       make([][]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func terraceDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TerraceInteractorIF, param TerraceWebInput, newDefault func(string) *TerraceWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ti.Draw())
	case "m", "move":
		return terraceMoveDispatch(bc, w, ti, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ti.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ti.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ti.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ti.Reset, ti.Hint, ti.ActionLog)
	}
	return true
}

func terraceMoveDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TerraceInteractorIF, param TerraceWebInput, newDefault func(string) *TerraceWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	// The terrace feeds the foundations and nothing else, so there is
	// deliberately no reserve -> tableau case here.
	case fromZone == "reserve" && toZone == "foundation":
		bc.writePresenterResponse(w, ti.MoveReserveToFoundation())
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, ti.MoveWasteToFoundation())
	case fromZone == "waste" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.MoveWasteToTableau(*param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.MoveTableauToTableau(*param.From.Col, *param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
