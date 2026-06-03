//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// YukonWebInput ユーコンWebインプット
type YukonWebInput struct {
	BaseWebInput
	From *YukonWebZone `json:"from,omitempty"`
	To   *YukonWebZone `json:"to,omitempty"`
}

// YukonWebZone ゾーン指定
type YukonWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// YukonWebOutputHint ヒント出力
type YukonWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// YukonWebOutput ユーコンWebアウトプット
type YukonWebOutput struct {
	Tableau    [][]*KlondikeWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                `json:"foundation"`
	Hint       *YukonWebOutputHint               `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// YukonWebController ユーコンWebコントローラークラス
type YukonWebController = GameWebController[usecase.YukonInteractorIF, YukonWebInput, *YukonWebOutput]

// NewYukonWebController and NewYukonWebControllerWithProvider are
// the standard and provider-backed constructors for YukonWebController.
var NewYukonWebController, NewYukonWebControllerWithProvider = webControllerPair[usecase.YukonInteractorIF, YukonWebInput, *YukonWebOutput](
	newYukonDefaultOutput, yukonDispatch,
)

func newYukonDefaultOutput(msg string) *YukonWebOutput {
	return &YukonWebOutput{
		Tableau:       make([][]*KlondikeWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func yukonDispatch(bc *baseController, w http.ResponseWriter, yi usecase.YukonInteractorIF, param YukonWebInput, newDefault func(string) *YukonWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return yukonMoveDispatch(bc, w, yi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, yi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, yi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, yi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, yi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, yi.Reset, yi.Hint, yi.ActionLog)
	}
	return true
}

func yukonMoveDispatch(bc *baseController, w http.ResponseWriter, yi usecase.YukonInteractorIF, param YukonWebInput, newDefault func(string) *YukonWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil, "param error: from.col, from.cardIndex, to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, yi.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, yi.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
