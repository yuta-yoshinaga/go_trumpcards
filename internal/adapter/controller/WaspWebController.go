//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WaspWebInput ワスプWebインプット
type WaspWebInput struct {
	BaseWebInput
	From *WaspWebZone `json:"from,omitempty"`
	To   *WaspWebZone `json:"to,omitempty"`
}

// WaspWebZone ゾーン指定
type WaspWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// WaspWebOutputHint ヒント出力
type WaspWebOutputHint struct {
	FromCol   int `json:"fromCol"`
	CardIndex int `json:"cardIndex"`
	ToCol     int `json:"toCol"`
}

// WaspWebOutput ワスプWebアウトプット
type WaspWebOutput struct {
	Tableau        [][]*KlondikeWebOutputTableauCard `json:"tableau"`
	StockCount     int                               `json:"stockCount"`
	CompletedSuits int                               `json:"completedSuits"`
	Hint           *WaspWebOutputHint                `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// WaspWebController ワスプWebコントローラークラス
type WaspWebController = GameWebController[usecase.WaspInteractorIF, WaspWebInput, *WaspWebOutput]

// NewWaspWebController and NewWaspWebControllerWithProvider are
// the standard and provider-backed constructors for WaspWebController.
var NewWaspWebController, NewWaspWebControllerWithProvider = webControllerPair[usecase.WaspInteractorIF, WaspWebInput, *WaspWebOutput](
	newWaspDefaultOutput, waspDispatch,
)

func newWaspDefaultOutput(msg string) *WaspWebOutput {
	return &WaspWebOutput{
		Tableau:       make([][]*KlondikeWebOutputTableauCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func waspDispatch(bc *baseController, w http.ResponseWriter, si usecase.WaspInteractorIF, param WaspWebInput, newDefault func(string) *WaspWebOutput) bool {
	switch param.Command {
	case "d", "deal":
		bc.writePresenterResponse(w, si.Deal())
	case "m", "move":
		return waspMoveDispatch(bc, w, si, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, si.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, si.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, si.Reset, si.Hint, si.ActionLog)
	}
	return true
}

func waspMoveDispatch(bc *baseController, w http.ResponseWriter, si usecase.WaspInteractorIF, param WaspWebInput, newDefault func(string) *WaspWebOutput) bool {
	mv := tableauMove{haveFrom: param.From != nil, haveTo: param.To != nil}
	if param.From != nil {
		mv.fromZone, mv.fromCol, mv.fromCardIndex = param.From.Zone, param.From.Col, param.From.CardIndex
	}
	if param.To != nil {
		mv.toZone, mv.toCol = param.To.Zone, param.To.Col
	}
	return dispatchTableauOnlyMove(bc, w, mv, si.MoveTableauToTableau, newDefault)
}
