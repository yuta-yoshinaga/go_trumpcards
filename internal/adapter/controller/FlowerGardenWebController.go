//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FlowerGardenWebInput Flower Garden Web インプット
type FlowerGardenWebInput struct {
	BaseWebInput
	From *FlowerGardenWebZone `json:"from,omitempty"`
	To   *FlowerGardenWebZone `json:"to,omitempty"`
}

// FlowerGardenWebZone ゾーン指定
type FlowerGardenWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// FlowerGardenWebOutputTableauCard タブローカード出力
type FlowerGardenWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// FlowerGardenWebOutputHint ヒント出力
type FlowerGardenWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// FlowerGardenWebOutput Flower Garden Web アウトプット
type FlowerGardenWebOutput struct {
	Tableau    [][]*FlowerGardenWebOutputTableauCard `json:"tableau"`
	Reserve    []*WebOutputCard                      `json:"reserve"`
	Foundation [][]*WebOutputCard                    `json:"foundation"`
	Hint       *FlowerGardenWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// FlowerGardenWebController Flower Garden Web コントローラークラス
type FlowerGardenWebController = GameWebController[usecase.FlowerGardenInteractorIF, FlowerGardenWebInput, *FlowerGardenWebOutput]

// NewFlowerGardenWebController and NewFlowerGardenWebControllerWithProvider are
// the standard and provider-backed constructors for FlowerGardenWebController.
var NewFlowerGardenWebController, NewFlowerGardenWebControllerWithProvider = webControllerPair[usecase.FlowerGardenInteractorIF, FlowerGardenWebInput, *FlowerGardenWebOutput](
	newFlowerGardenDefaultOutput, flowerGardenDispatch,
)

func newFlowerGardenDefaultOutput(msg string) *FlowerGardenWebOutput {
	return &FlowerGardenWebOutput{
		Tableau:       make([][]*FlowerGardenWebOutputTableauCard, 0),
		Reserve:       make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func flowerGardenDispatch(bc *baseController, w http.ResponseWriter, bi usecase.FlowerGardenInteractorIF, param FlowerGardenWebInput, newDefault func(string) *FlowerGardenWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return flowerGardenMoveDispatch(bc, w, bi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, bi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, bi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, bi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, bi.Reset, bi.Hint, bi.ActionLog)
	}
	return true
}

func flowerGardenMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.FlowerGardenInteractorIF, param FlowerGardenWebInput, newDefault func(string) *FlowerGardenWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		// Flower Garden only ever moves the end card; pass -1 so the domain
		// resolves the index from its own state.
		bc.writePresenterResponse(w, bi.MoveTableauToTableau(*param.From.Col, -1, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "reserve" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveReserveToTableau(*param.From.Col, *param.To.Col))
	case fromZone == "reserve" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveReserveToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
