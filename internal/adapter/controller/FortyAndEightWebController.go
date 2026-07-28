//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FortyAndEightWebInput フォーティ・アンド・エイトWebインプット
type FortyAndEightWebInput struct {
	BaseWebInput
	From *FortyAndEightWebZone `json:"from,omitempty"`
	To   *FortyAndEightWebZone `json:"to,omitempty"`
}

// FortyAndEightWebZone ゾーン指定
type FortyAndEightWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// FortyAndEightWebOutputTableauCard タブローカード出力
type FortyAndEightWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// FortyAndEightWebOutputHint ヒント出力
type FortyAndEightWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// FortyAndEightWebOutput フォーティ・アンド・エイトWebアウトプット
type FortyAndEightWebOutput struct {
	Tableau    [][]*FortyAndEightWebOutputTableauCard `json:"tableau"`
	StockCount int                                    `json:"stockCount"`
	Waste      []*WebOutputCard                       `json:"waste"`
	Foundation [][]*WebOutputCard                     `json:"foundation"`
	RedealUsed bool                                   `json:"redealUsed"`
	CanRedeal  bool                                   `json:"canRedeal"`
	Hint       *FortyAndEightWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// FortyAndEightWebController フォーティ・アンド・エイトWebコントローラークラス
type FortyAndEightWebController = GameWebController[usecase.FortyAndEightInteractorIF, FortyAndEightWebInput, *FortyAndEightWebOutput]

// NewFortyAndEightWebController and NewFortyAndEightWebControllerWithProvider are
// the standard and provider-backed constructors for FortyAndEightWebController.
var NewFortyAndEightWebController, NewFortyAndEightWebControllerWithProvider = webControllerPair[usecase.FortyAndEightInteractorIF, FortyAndEightWebInput, *FortyAndEightWebOutput](
	newFortyAndEightDefaultOutput, fortyAndEightDispatch,
)

func newFortyAndEightDefaultOutput(msg string) *FortyAndEightWebOutput {
	return &FortyAndEightWebOutput{
		Tableau:       make([][]*FortyAndEightWebOutputTableauCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func fortyAndEightDispatch(bc *baseController, w http.ResponseWriter, fi usecase.FortyAndEightInteractorIF, param FortyAndEightWebInput, newDefault func(string) *FortyAndEightWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, fi.Draw())
	case "rd", "redeal":
		bc.writePresenterResponse(w, fi.Redeal())
	case "m", "move":
		return fortyAndEightMoveDispatch(bc, w, fi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, fi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, fi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, fi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, fi.Reset, fi.Hint, fi.ActionLog)
	}
	return true
}

func fortyAndEightMoveDispatch(bc *baseController, w http.ResponseWriter, fi usecase.FortyAndEightInteractorIF, param FortyAndEightWebInput, newDefault func(string) *FortyAndEightWebOutput) bool {
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
		bc.writePresenterResponse(w, fi.MoveWasteToTableau(*param.To.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, fi.MoveWasteToFoundation())
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil, "param error: from.col, from.cardIndex, to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
