//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FourSeasonsWebInput フォーシーズンズWebインプット
type FourSeasonsWebInput struct {
	BaseWebInput
	From *FourSeasonsWebZone `json:"from,omitempty"`
	To   *FourSeasonsWebZone `json:"to,omitempty"`
}

// FourSeasonsWebZone ゾーン指定。
// タブローは最上段の1枚しか動かないので cardIndex は無い（Canfield との違い）。
type FourSeasonsWebZone struct {
	Zone string `json:"zone"`
	Idx  *int   `json:"idx,omitempty"`
}

// FourSeasonsWebOutputHint ヒント出力
type FourSeasonsWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// FourSeasonsWebOutput フォーシーズンズWebアウトプット
type FourSeasonsWebOutput struct {
	Tableau    [][]*WebOutputCard        `json:"tableau"`
	Foundation [][]*WebOutputCard        `json:"foundation"`
	StockCount int                       `json:"stockCount"`
	Waste      []*WebOutputCard          `json:"waste"`
	BaseRank   int                       `json:"baseRank"`
	Phase      int                       `json:"phase"`
	MoveCount  int                       `json:"moveCount"`
	CanUndo    bool                      `json:"canUndo"`
	Hint       *FourSeasonsWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// FourSeasonsWebController フォーシーズンズWebコントローラークラス
type FourSeasonsWebController = GameWebController[usecase.FourSeasonsInteractorIF, FourSeasonsWebInput, *FourSeasonsWebOutput]

// NewFourSeasonsWebController and NewFourSeasonsWebControllerWithProvider are
// the standard and provider-backed constructors for FourSeasonsWebController.
var NewFourSeasonsWebController, NewFourSeasonsWebControllerWithProvider = webControllerPair[usecase.FourSeasonsInteractorIF, FourSeasonsWebInput, *FourSeasonsWebOutput](
	newFourSeasonsDefaultOutput, fourSeasonsDispatch,
)

func newFourSeasonsDefaultOutput(msg string) *FourSeasonsWebOutput {
	return &FourSeasonsWebOutput{
		Tableau:       make([][]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func fourSeasonsDispatch(bc *baseController, w http.ResponseWriter, ci usecase.FourSeasonsInteractorIF, param FourSeasonsWebInput, newDefault func(string) *FourSeasonsWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "m", "move":
		return fourSeasonsMoveDispatch(bc, w, ci, param, newDefault)
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

// fourSeasonsMoveDispatch は4通りの移動を捌く。
// **ファンデーションは必ず idx を要求する。** 四隅のどこに載るかはベースランクで
// 開いた順に決まるので、Canfield のように「1つだけ」と決め打ちできない。
func fourSeasonsMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.FourSeasonsInteractorIF, param FourSeasonsWebInput, newDefault func(string) *FourSeasonsWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "waste" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Idx == nil, "param error: to.idx is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveWasteToTableau(*param.To.Idx))
	case fromZone == "waste" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.To.Idx == nil, "param error: to.idx is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveWasteToFoundation(*param.To.Idx))
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Idx == nil || param.To.Idx == nil, "param error: from.idx and to.idx are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToTableau(*param.From.Idx, *param.To.Idx))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Idx == nil || param.To.Idx == nil, "param error: from.idx and to.idx are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToFoundation(*param.From.Idx, *param.To.Idx))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
