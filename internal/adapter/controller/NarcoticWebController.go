//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NarcoticWebInput ナルコティックWebインプット
type NarcoticWebInput struct {
	BaseWebInput
	Col *int `json:"col,omitempty"`
}

// NarcoticWebOutputCard 場札カード出力
type NarcoticWebOutputCard struct {
	Card      *WebOutputCard `json:"card"`
	Top       bool           `json:"top"`
	Removable bool           `json:"removable"`
	Movable   bool           `json:"movable"`
}

// NarcoticWebOutputHint ヒント出力
type NarcoticWebOutputHint struct {
	Type string `json:"type"`
	Col  int    `json:"col"`
}

// NarcoticWebOutput ナルコティックWebアウトプット
type NarcoticWebOutput struct {
	Columns      [][]*NarcoticWebOutputCard `json:"columns"`
	StockCount   int                        `json:"stockCount"`
	DiscardCount int                        `json:"discardCount"`
	// DiscardTop は直近に除去した札（捨て札パイルの一番上）。捨て札が空なら省略。
	DiscardTop *WebOutputCard         `json:"discardTop,omitempty"`
	Hint       *NarcoticWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// NarcoticWebController ナルコティックWebコントローラークラス
type NarcoticWebController = GameWebController[usecase.NarcoticInteractorIF, NarcoticWebInput, *NarcoticWebOutput]

// NewNarcoticWebController and NewNarcoticWebControllerWithProvider are
// the standard and provider-backed constructors for NarcoticWebController.
var NewNarcoticWebController, NewNarcoticWebControllerWithProvider = webControllerPair[usecase.NarcoticInteractorIF, NarcoticWebInput, *NarcoticWebOutput](
	newNarcoticDefaultOutput, narcoticDispatch,
)

func newNarcoticDefaultOutput(msg string) *NarcoticWebOutput {
	return &NarcoticWebOutput{
		Columns:       make([][]*NarcoticWebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func narcoticDispatch(bc *baseController, w http.ResponseWriter, ai usecase.NarcoticInteractorIF, param NarcoticWebInput, newDefault func(string) *NarcoticWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ai.Draw())
	case "rm", "remove":
		// **col は要らない。**揃った4枚をまとめて捨てるので、対象は盤面から一意。
		bc.writePresenterResponse(w, ai.Remove())
	case "rd", "redeal":
		bc.writePresenterResponse(w, ai.Redeal())
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
