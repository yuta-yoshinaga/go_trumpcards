//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AcesUpWebInput エースアップWebインプット
type AcesUpWebInput struct {
	BaseWebInput
	Col *int `json:"col,omitempty"`
}

// AcesUpWebOutputCard 場札カード出力
type AcesUpWebOutputCard struct {
	Card      *WebOutputCard `json:"card"`
	Top       bool           `json:"top"`
	Removable bool           `json:"removable"`
	Movable   bool           `json:"movable"`
}

// AcesUpWebOutputHint ヒント出力
type AcesUpWebOutputHint struct {
	Type string `json:"type"`
	Col  int    `json:"col"`
}

// AcesUpWebOutput エースアップWebアウトプット
type AcesUpWebOutput struct {
	Columns      [][]*AcesUpWebOutputCard `json:"columns"`
	StockCount   int                      `json:"stockCount"`
	DiscardCount int                      `json:"discardCount"`
	// DiscardTop は直近に除去した札（捨て札パイルの一番上）。捨て札が空なら省略。
	DiscardTop *WebOutputCard       `json:"discardTop,omitempty"`
	Hint       *AcesUpWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// AcesUpWebController エースアップWebコントローラークラス
type AcesUpWebController = GameWebController[usecase.AcesUpInteractorIF, AcesUpWebInput, *AcesUpWebOutput]

// NewAcesUpWebController and NewAcesUpWebControllerWithProvider are
// the standard and provider-backed constructors for AcesUpWebController.
var NewAcesUpWebController, NewAcesUpWebControllerWithProvider = webControllerPair[usecase.AcesUpInteractorIF, AcesUpWebInput, *AcesUpWebOutput](
	newAcesUpDefaultOutput, acesUpDispatch,
)

func newAcesUpDefaultOutput(msg string) *AcesUpWebOutput {
	return &AcesUpWebOutput{
		Columns:       make([][]*AcesUpWebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func acesUpDispatch(bc *baseController, w http.ResponseWriter, ai usecase.AcesUpInteractorIF, param AcesUpWebInput, newDefault func(string) *AcesUpWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ai.Draw())
	case "rm", "remove":
		if !requireParam(bc, w, newDefault, param.Col == nil, "param error: col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.Remove(*param.Col))
	case "mv", "move":
		if !requireParam(bc, w, newDefault, param.Col == nil, "param error: col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.Move(*param.Col))
	case "g", "giveup":
		bc.writePresenterResponse(w, ai.GiveUp())
	case "u", "undo":
		bc.writePresenterResponse(w, ai.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ai.Reset, ai.Hint, ai.ActionLog)
	}
	return true
}
