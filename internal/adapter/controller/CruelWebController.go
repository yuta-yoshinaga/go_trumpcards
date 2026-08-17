//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CruelWebInput クルーエルWebインプット
type CruelWebInput struct {
	BaseWebInput
	From *CruelWebZone `json:"from,omitempty"`
	To   *CruelWebZone `json:"to,omitempty"`
}

// CruelWebZone ゾーン指定
type CruelWebZone struct {
	Zone string `json:"zone"`
	Col  *int   `json:"col,omitempty"`
}

// CruelWebOutputHint ヒント出力
type CruelWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// CruelWebOutput クルーエルWebアウトプット
type CruelWebOutput struct {
	Tableau    [][]*KlondikeWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                `json:"foundation"`
	Hint       *CruelWebOutputHint               `json:"hint,omitempty"`
	// CanAutoComplete は今オートコンプリートで動かせる札があるか。ボタンの
	// 有効/無効に使う。判定は AutoComplete と同じものを domain で共有している
	// ので、フロントで組札の中身から推測しない (#5496)。
	CanAutoComplete bool `json:"canAutoComplete"`
	SolitaireWebOutputBase
	WebOutputBase
}

// CruelWebController クルーエルWebコントローラークラス
type CruelWebController = GameWebController[usecase.CruelInteractorIF, CruelWebInput, *CruelWebOutput]

// NewCruelWebController and NewCruelWebControllerWithProvider are the standard
// and provider-backed constructors for CruelWebController.
var NewCruelWebController, NewCruelWebControllerWithProvider = webControllerPair[usecase.CruelInteractorIF, CruelWebInput, *CruelWebOutput](
	newCruelDefaultOutput, cruelDispatch,
)

func newCruelDefaultOutput(msg string) *CruelWebOutput {
	return &CruelWebOutput{
		Tableau:       make([][]*KlondikeWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func cruelDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CruelInteractorIF, param CruelWebInput, newDefault func(string) *CruelWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return cruelMoveDispatch(bc, w, ci, param, newDefault)
	case "s", "shift":
		bc.writePresenterResponse(w, ci.Shift())
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

func cruelMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CruelInteractorIF, param CruelWebInput, newDefault func(string) *CruelWebOutput) bool {
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
		bc.writePresenterResponse(w, ci.MoveTableauToTableau(*param.From.Col, *param.To.Col))
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
