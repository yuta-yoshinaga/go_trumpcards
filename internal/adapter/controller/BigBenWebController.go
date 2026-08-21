//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BigBenWebInput ビッグ・ベン Web インプット
type BigBenWebInput struct {
	BaseWebInput
	From *BigBenWebZone `json:"from,omitempty"`
	To   *BigBenWebZone `json:"to,omitempty"`
}

// BigBenWebZone ゾーン指定。Zone は "tableau" / "foundation"。
type BigBenWebZone struct {
	Zone string `json:"zone"`
	Col  *int   `json:"col,omitempty"`
}

// BigBenWebOutputTableauCard タブローカード出力
type BigBenWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// BigBenWebOutputFoundation 文字盤 1 つぶんの出力
type BigBenWebOutputFoundation struct {
	Cards []*WebOutputCard `json:"cards"`
	// TargetRank この文字盤が到達すべきランク（1 時が A の 1、12 時が Q の 12）
	TargetRank int `json:"targetRank"`
	// Complete 目標ランクに達しているか
	Complete bool `json:"complete"`
}

// BigBenWebOutputHint ヒント出力
type BigBenWebOutputHint struct {
	FromCol int    `json:"fromCol"`
	ToZone  string `json:"toZone"`
	ToIdx   int    `json:"toIdx"`
}

// BigBenWebOutput ビッグ・ベン Web アウトプット
type BigBenWebOutput struct {
	Tableau    [][]*BigBenWebOutputTableauCard `json:"tableau"`
	Foundation []*BigBenWebOutputFoundation    `json:"foundation"`
	// StockCount 山札の残り枚数。**UI はこれを見て補充ボタンを無効化する。**
	// 見ないと、空の山札を押してサーバに拒まれるまで気付けない。
	StockCount int                  `json:"stockCount"`
	Hint       *BigBenWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// BigBenWebController ビッグ・ベン Web コントローラークラス
type BigBenWebController = GameWebController[usecase.BigBenInteractorIF, BigBenWebInput, *BigBenWebOutput]

// NewBigBenWebController and NewBigBenWebControllerWithProvider are
// the standard and provider-backed constructors for BigBenWebController.
var NewBigBenWebController, NewBigBenWebControllerWithProvider = webControllerPair[usecase.BigBenInteractorIF, BigBenWebInput, *BigBenWebOutput](
	newBigBenDefaultOutput, bigBenDispatch,
)

func newBigBenDefaultOutput(msg string) *BigBenWebOutput {
	return &BigBenWebOutput{
		Tableau:       make([][]*BigBenWebOutputTableauCard, 0),
		Foundation:    make([]*BigBenWebOutputFoundation, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func bigBenDispatch(bc *baseController, w http.ResponseWriter, gi usecase.BigBenInteractorIF, param BigBenWebInput, newDefault func(string) *BigBenWebOutput) bool {
	switch param.Command {
	case "d", "deal":
		bc.writePresenterResponse(w, gi.Deal())
	case "m", "move":
		return bigBenMoveDispatch(bc, w, gi, param, newDefault)
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

func bigBenMoveDispatch(bc *baseController, w http.ResponseWriter, gi usecase.BigBenInteractorIF, param BigBenWebInput, newDefault func(string) *BigBenWebOutput) bool {
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
