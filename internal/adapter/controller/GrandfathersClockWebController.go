//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GrandfathersClockWebInput グランドファーザーズ・クロック Web インプット
type GrandfathersClockWebInput struct {
	BaseWebInput
	From *GrandfathersClockWebZone `json:"from,omitempty"`
	To   *GrandfathersClockWebZone `json:"to,omitempty"`
}

// GrandfathersClockWebZone ゾーン指定。Zone は "tableau" / "foundation"。
type GrandfathersClockWebZone struct {
	Zone string `json:"zone"`
	Col  *int   `json:"col,omitempty"`
}

// GrandfathersClockWebOutputTableauCard タブローカード出力
type GrandfathersClockWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// GrandfathersClockWebOutputFoundation 文字盤 1 つぶんの出力
type GrandfathersClockWebOutputFoundation struct {
	Cards []*WebOutputCard `json:"cards"`
	// TargetRank この文字盤が到達すべきランク（1 時が A の 1、12 時が Q の 12）
	TargetRank int `json:"targetRank"`
	// Complete 目標ランクに達しているか
	Complete bool `json:"complete"`
}

// GrandfathersClockWebOutputHint ヒント出力
type GrandfathersClockWebOutputHint struct {
	FromCol int    `json:"fromCol"`
	ToZone  string `json:"toZone"`
	ToIdx   int    `json:"toIdx"`
}

// GrandfathersClockWebOutput グランドファーザーズ・クロック Web アウトプット
type GrandfathersClockWebOutput struct {
	Tableau    [][]*GrandfathersClockWebOutputTableauCard `json:"tableau"`
	Foundation []*GrandfathersClockWebOutputFoundation    `json:"foundation"`
	Hint       *GrandfathersClockWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// GrandfathersClockWebController グランドファーザーズ・クロック Web コントローラークラス
type GrandfathersClockWebController = GameWebController[usecase.GrandfathersClockInteractorIF, GrandfathersClockWebInput, *GrandfathersClockWebOutput]

// NewGrandfathersClockWebController and NewGrandfathersClockWebControllerWithProvider are
// the standard and provider-backed constructors for GrandfathersClockWebController.
var NewGrandfathersClockWebController, NewGrandfathersClockWebControllerWithProvider = webControllerPair[usecase.GrandfathersClockInteractorIF, GrandfathersClockWebInput, *GrandfathersClockWebOutput](
	newGrandfathersClockDefaultOutput, grandfathersClockDispatch,
)

func newGrandfathersClockDefaultOutput(msg string) *GrandfathersClockWebOutput {
	return &GrandfathersClockWebOutput{
		Tableau:       make([][]*GrandfathersClockWebOutputTableauCard, 0),
		Foundation:    make([]*GrandfathersClockWebOutputFoundation, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func grandfathersClockDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GrandfathersClockInteractorIF, param GrandfathersClockWebInput, newDefault func(string) *GrandfathersClockWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return grandfathersClockMoveDispatch(bc, w, gi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, gi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, gi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, gi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, gi.Reset, gi.Hint, gi.ActionLog)
	}
	return true
}

func grandfathersClockMoveDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GrandfathersClockInteractorIF, param GrandfathersClockWebInput, newDefault func(string) *GrandfathersClockWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	if !requireParam(bc, w, newDefault, param.From.Zone != "tableau" || param.From.Col == nil, "param error: from.zone must be tableau with a col.") {
		return true
	}

	switch param.To.Zone {
	case "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.MoveTableauToTableau(*param.From.Col, *param.To.Col))
	case "foundation":
		// 文字盤は 12 個あり、同じスートの札が複数の文字盤に載りうるので、
		// どの文字盤かはクライアントが指定する必要がある（Bisley のように
		// スートから導けない）。
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.MoveTableauToFoundation(*param.From.Col, *param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
