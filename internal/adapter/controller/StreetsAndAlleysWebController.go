//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// StreetsAndAlleysWebInput Streets and Alleys Web インプット
type StreetsAndAlleysWebInput struct {
	BaseWebInput
	From *StreetsAndAlleysWebZone `json:"from,omitempty"`
	To   *StreetsAndAlleysWebZone `json:"to,omitempty"`
}

// StreetsAndAlleysWebZone ゾーン指定
type StreetsAndAlleysWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// StreetsAndAlleysWebOutputTableauCard タブローカード出力
type StreetsAndAlleysWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// StreetsAndAlleysWebOutputHint ヒント出力
type StreetsAndAlleysWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// StreetsAndAlleysWebOutput Streets and Alleys Web アウトプット
type StreetsAndAlleysWebOutput struct {
	Tableau    [][]*StreetsAndAlleysWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                        `json:"foundation"`
	Hint       *StreetsAndAlleysWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// StreetsAndAlleysWebController Streets and Alleys Web コントローラークラス
type StreetsAndAlleysWebController = GameWebController[usecase.StreetsAndAlleysInteractorIF, StreetsAndAlleysWebInput, *StreetsAndAlleysWebOutput]

// NewStreetsAndAlleysWebController and NewStreetsAndAlleysWebControllerWithProvider are
// the standard and provider-backed constructors for StreetsAndAlleysWebController.
var NewStreetsAndAlleysWebController, NewStreetsAndAlleysWebControllerWithProvider = webControllerPair[usecase.StreetsAndAlleysInteractorIF, StreetsAndAlleysWebInput, *StreetsAndAlleysWebOutput](
	newStreetsAndAlleysDefaultOutput, streetsAndAlleysDispatch,
)

func newStreetsAndAlleysDefaultOutput(msg string) *StreetsAndAlleysWebOutput {
	return &StreetsAndAlleysWebOutput{
		Tableau:       make([][]*StreetsAndAlleysWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func streetsAndAlleysDispatch(bc *baseController, w http.ResponseWriter, bi usecase.StreetsAndAlleysInteractorIF, param StreetsAndAlleysWebInput, newDefault func(string) *StreetsAndAlleysWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return streetsAndAlleysMoveDispatch(bc, w, bi, param, newDefault)
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

func streetsAndAlleysMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.StreetsAndAlleysInteractorIF, param StreetsAndAlleysWebInput, newDefault func(string) *StreetsAndAlleysWebOutput) bool {
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
		// Streets and Alleys only ever moves the top card; pass -1 so the
		// domain resolves the index from its own state.
		bc.writePresenterResponse(w, bi.MoveTableauToTableau(*param.From.Col, -1, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
