//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LaBelleLucieWebInput ラ・ベル・ルーシーのWebインプット。
type LaBelleLucieWebInput struct {
	BaseWebInput
	// From 移動元の扇番号。
	From *int `json:"from,omitempty"`
	// To 移動先の扇番号 (mf のとき)。
	To *int `json:"to,omitempty"`
}

// LaBelleLucieWebOutputHint ヒント出力。
type LaBelleLucieWebOutputHint struct {
	FromFan      int  `json:"fromFan"`
	ToFan        int  `json:"toFan"`
	ToFoundation bool `json:"toFoundation"`
}

// LaBelleLucieWebOutput ラ・ベル・ルーシーのWebアウトプット。
type LaBelleLucieWebOutput struct {
	Fans        [][]*WebOutputCard         `json:"fans"`
	Foundation  [][]*WebOutputCard         `json:"foundation"`
	RedealsLeft int                        `json:"redealsLeft"`
	Phase       int                        `json:"phase"`
	MoveCount   int                        `json:"moveCount"`
	CanUndo     bool                       `json:"canUndo"`
	Hint        *LaBelleLucieWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// LaBelleLucieWebController ラ・ベル・ルーシーのWebコントローラークラス。
type LaBelleLucieWebController = GameWebController[usecase.LaBelleLucieInteractorIF, LaBelleLucieWebInput, *LaBelleLucieWebOutput]

// NewLaBelleLucieWebController and NewLaBelleLucieWebControllerWithProvider are
// the standard and provider-backed constructors for LaBelleLucieWebController.
var NewLaBelleLucieWebController, NewLaBelleLucieWebControllerWithProvider = webControllerPair[usecase.LaBelleLucieInteractorIF, LaBelleLucieWebInput, *LaBelleLucieWebOutput](
	newLaBelleLucieDefaultOutput, laBelleLucieDispatch,
)

func newLaBelleLucieDefaultOutput(msg string) *LaBelleLucieWebOutput {
	return &LaBelleLucieWebOutput{
		Fans:          make([][]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func laBelleLucieDispatch(bc *baseController, w http.ResponseWriter, li usecase.LaBelleLucieInteractorIF, param LaBelleLucieWebInput, newDefault func(string) *LaBelleLucieWebOutput) bool {
	switch param.Command {
	case "mf":
		if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
			return true
		}
		bc.writePresenterResponse(w, li.MoveFanToFan(*param.From, *param.To))
	case "ff":
		if !requireParam(bc, w, newDefault, param.From == nil, "param error: from is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.MoveFanToFoundation(*param.From))
	case "rd", "redeal":
		bc.writePresenterResponse(w, li.Redeal())
	case "g", "giveup":
		bc.writePresenterResponse(w, li.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, li.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, li.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, li.Reset, li.Hint, li.ActionLog)
	}
	return true
}
