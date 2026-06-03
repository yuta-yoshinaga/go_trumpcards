//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CanfieldWebInput キャンフィールドWebインプット
type CanfieldWebInput struct {
	BaseWebInput
	From *CanfieldWebZone `json:"from,omitempty"`
	To   *CanfieldWebZone `json:"to,omitempty"`
}

// CanfieldWebZone ゾーン指定
type CanfieldWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// CanfieldWebOutputTableauCard タブローカード出力
type CanfieldWebOutputTableauCard struct {
	Card *WebOutputCard `json:"card"`
}

// CanfieldWebOutputHint ヒント出力
type CanfieldWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// CanfieldWebOutput キャンフィールドWebアウトプット
type CanfieldWebOutput struct {
	Tableau    [][]*CanfieldWebOutputTableauCard `json:"tableau"`
	Reserve    []*WebOutputCard                  `json:"reserve"`
	StockCount int                               `json:"stockCount"`
	Waste      []*WebOutputCard                  `json:"waste"`
	Foundation [][]*WebOutputCard                `json:"foundation"`
	BaseRank   int                               `json:"baseRank"`
	Phase      int                               `json:"phase"`
	MoveCount  int                               `json:"moveCount"`
	CanUndo    bool                              `json:"canUndo"`
	Hint       *CanfieldWebOutputHint            `json:"hint,omitempty"`
	WebOutputBase
}

// CanfieldWebController キャンフィールドWebコントローラークラス
type CanfieldWebController = GameWebController[usecase.CanfieldInteractorIF, CanfieldWebInput, *CanfieldWebOutput]

// NewCanfieldWebController and NewCanfieldWebControllerWithProvider are
// the standard and provider-backed constructors for CanfieldWebController.
var NewCanfieldWebController, NewCanfieldWebControllerWithProvider = webControllerPair[usecase.CanfieldInteractorIF, CanfieldWebInput, *CanfieldWebOutput](
	newCanfieldDefaultOutput, canfieldDispatch,
)

func newCanfieldDefaultOutput(msg string) *CanfieldWebOutput {
	return &CanfieldWebOutput{
		Tableau:       make([][]*CanfieldWebOutputTableauCard, 0),
		Reserve:       make([]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func canfieldDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CanfieldInteractorIF, param CanfieldWebInput, newDefault func(string) *CanfieldWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "m", "move":
		return canfieldMoveDispatch(bc, w, ci, param, newDefault)
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

func canfieldMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CanfieldInteractorIF, param CanfieldWebInput, newDefault func(string) *CanfieldWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "waste" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveWasteToTableau(*param.To.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, ci.MoveWasteToFoundation())
	case fromZone == "reserve" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveReserveToTableau(*param.To.Col))
	case fromZone == "reserve" && toZone == "foundation":
		bc.writePresenterResponse(w, ci.MoveReserveToFoundation())
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil, "param error: from.col, from.cardIndex, to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
