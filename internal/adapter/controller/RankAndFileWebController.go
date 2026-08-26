//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RankAndFileWebInput ランク・アンド・ファイルWebインプット
type RankAndFileWebInput struct {
	BaseWebInput
	From *RankAndFileWebZone `json:"from,omitempty"`
	To   *RankAndFileWebZone `json:"to,omitempty"`
}

// RankAndFileWebZone ゾーン指定
type RankAndFileWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// RankAndFileWebOutputTableauCard タブローカード出力
type RankAndFileWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// RankAndFileWebOutputHint ヒント出力
type RankAndFileWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// RankAndFileWebOutput ランク・アンド・ファイルWebアウトプット
type RankAndFileWebOutput struct {
	Tableau    [][]*RankAndFileWebOutputTableauCard `json:"tableau"`
	StockCount int                                  `json:"stockCount"`
	Waste      []*WebOutputCard                     `json:"waste"`
	Foundation [][]*WebOutputCard                   `json:"foundation"`
	Hint       *RankAndFileWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// RankAndFileWebController ランク・アンド・ファイルWebコントローラークラス
type RankAndFileWebController = GameWebController[usecase.RankAndFileInteractorIF, RankAndFileWebInput, *RankAndFileWebOutput]

// NewRankAndFileWebController and NewRankAndFileWebControllerWithProvider are
// the standard and provider-backed constructors for RankAndFileWebController.
var NewRankAndFileWebController, NewRankAndFileWebControllerWithProvider = webControllerPair[usecase.RankAndFileInteractorIF, RankAndFileWebInput, *RankAndFileWebOutput](
	newRankAndFileDefaultOutput, rankAndFileDispatch,
)

func newRankAndFileDefaultOutput(msg string) *RankAndFileWebOutput {
	return &RankAndFileWebOutput{
		Tableau:       make([][]*RankAndFileWebOutputTableauCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func rankAndFileDispatch(bc *baseController, w http.ResponseWriter, fi usecase.RankAndFileInteractorIF, param RankAndFileWebInput, newDefault func(string) *RankAndFileWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, fi.Draw())
	case "m", "move":
		return rankAndFileMoveDispatch(bc, w, fi, param, newDefault)
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

func rankAndFileMoveDispatch(bc *baseController, w http.ResponseWriter, fi usecase.RankAndFileInteractorIF, param RankAndFileWebInput, newDefault func(string) *RankAndFileWebOutput) bool {
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
