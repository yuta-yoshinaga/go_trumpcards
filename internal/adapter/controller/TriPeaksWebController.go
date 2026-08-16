//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TriPeaksWebInput トリピークスWebインプット
type TriPeaksWebInput struct {
	BaseWebInput
	Row *int `json:"row,omitempty"`
	Col *int `json:"col,omitempty"`
}

// TriPeaksWebOutputCard タブローカード出力
type TriPeaksWebOutputCard struct {
	Card    *WebOutputCard `json:"card"`
	Removed bool           `json:"removed"`
	Exposed bool           `json:"exposed"`
}

// TriPeaksWebOutputHint ヒント出力
type TriPeaksWebOutputHint struct {
	Type string `json:"type"`
	Row  int    `json:"row"`
	Col  int    `json:"col"`
}

// TriPeaksWebOutput トリピークスWebアウトプット
type TriPeaksWebOutput struct {
	Layout     [][]*TriPeaksWebOutputCard `json:"layout"`
	StockCount int                        `json:"stockCount"`
	Waste      []*WebOutputCard           `json:"waste"`
	Hint       *TriPeaksWebOutputHint     `json:"hint,omitempty"`
	// Score / Combo はドメインが数える。以前は得点計算そのものがフロントの
	// useTriPeaksScore にしか無く、CUI からは参照する値が存在しなかった (#5511)。
	Score int `json:"score"`
	Combo int `json:"combo"`
	SolitaireWebOutputBase
	WebOutputBase
}

// TriPeaksWebController トリピークスWebコントローラークラス
type TriPeaksWebController = GameWebController[usecase.TriPeaksInteractorIF, TriPeaksWebInput, *TriPeaksWebOutput]

// NewTriPeaksWebController and NewTriPeaksWebControllerWithProvider are
// the standard and provider-backed constructors for TriPeaksWebController.
var NewTriPeaksWebController, NewTriPeaksWebControllerWithProvider = webControllerPair[usecase.TriPeaksInteractorIF, TriPeaksWebInput, *TriPeaksWebOutput](
	newTriPeaksDefaultOutput, triPeaksDispatch,
)

func newTriPeaksDefaultOutput(msg string) *TriPeaksWebOutput {
	return &TriPeaksWebOutput{
		Layout:        make([][]*TriPeaksWebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func triPeaksDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TriPeaksInteractorIF, param TriPeaksWebInput, newDefault func(string) *TriPeaksWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ti.Draw())
	case "rm", "remove":
		if !requireParam(bc, w, newDefault, param.Row == nil || param.Col == nil, "param error: row and col are required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Remove(*param.Row, *param.Col))
	case "g", "giveup":
		bc.writePresenterResponse(w, ti.GiveUp())
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
