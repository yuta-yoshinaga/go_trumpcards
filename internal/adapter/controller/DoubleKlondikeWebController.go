//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DoubleKlondikeWebInput ダブル・クロンダイクのWebインプット。
type DoubleKlondikeWebInput struct {
	BaseWebInput
	// Col 単一列指定 (mwt / mtf)。
	Col *int `json:"col,omitempty"`
	// FromCol / CardIndex / ToCol タブロー間移動 (mtt)。
	FromCol   *int `json:"fromCol,omitempty"`
	CardIndex *int `json:"cardIndex,omitempty"`
	ToCol     *int `json:"toCol,omitempty"`
}

// DoubleKlondikeWebOutputTableauCard タブローカード出力。裏向きは card を隠す。
type DoubleKlondikeWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// DoubleKlondikeWebOutputHint ヒント出力。
type DoubleKlondikeWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// DoubleKlondikeWebOutput ダブル・クロンダイクのWebアウトプット。
type DoubleKlondikeWebOutput struct {
	Tableau     [][]*DoubleKlondikeWebOutputTableauCard `json:"tableau"`
	StockCount  int                                     `json:"stockCount"`
	Waste       []*WebOutputCard                        `json:"waste"`
	Foundation  [][]*WebOutputCard                      `json:"foundation"`
	Phase       int                                     `json:"phase"`
	MoveCount   int                                     `json:"moveCount"`
	CanUndo     bool                                    `json:"canUndo"`
	IsStalemate bool                                    `json:"isStalemate"`
	Hint        *DoubleKlondikeWebOutputHint            `json:"hint,omitempty"`
	WebOutputBase
}

// DoubleKlondikeWebController ダブル・クロンダイクのWebコントローラークラス。
type DoubleKlondikeWebController = GameWebController[usecase.DoubleKlondikeInteractorIF, DoubleKlondikeWebInput, *DoubleKlondikeWebOutput]

// NewDoubleKlondikeWebController and NewDoubleKlondikeWebControllerWithProvider are the
// standard and provider-backed constructors for DoubleKlondikeWebController.
var NewDoubleKlondikeWebController, NewDoubleKlondikeWebControllerWithProvider = webControllerPair[usecase.DoubleKlondikeInteractorIF, DoubleKlondikeWebInput, *DoubleKlondikeWebOutput](
	newDoubleKlondikeDefaultOutput, doubleKlondikeDispatch,
)

func newDoubleKlondikeDefaultOutput(msg string) *DoubleKlondikeWebOutput {
	return &DoubleKlondikeWebOutput{
		Tableau:       make([][]*DoubleKlondikeWebOutputTableauCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func doubleKlondikeDispatch(bc *baseController, w http.ResponseWriter, di usecase.DoubleKlondikeInteractorIF, param DoubleKlondikeWebInput, newDefault func(string) *DoubleKlondikeWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, di.Draw())
	case "mwt":
		if !requireParam(bc, w, newDefault, param.Col == nil, "param error: col is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.MoveWasteToTableau(*param.Col))
	case "mwf":
		bc.writePresenterResponse(w, di.MoveWasteToFoundation())
	case "mtf":
		if !requireParam(bc, w, newDefault, param.Col == nil, "param error: col is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.MoveTableauToFoundation(*param.Col))
	case "mtt":
		if !requireParam(bc, w, newDefault, param.FromCol == nil || param.CardIndex == nil || param.ToCol == nil, "param error: fromCol, cardIndex and toCol are required.") {
			return true
		}
		bc.writePresenterResponse(w, di.MoveTableauToTableau(*param.FromCol, *param.CardIndex, *param.ToCol))
	case "g", "giveup":
		bc.writePresenterResponse(w, di.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, di.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, di.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, di.Reset, di.Hint, di.ActionLog)
	}
	return true
}
