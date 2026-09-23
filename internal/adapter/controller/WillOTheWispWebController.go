//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WillOTheWispWebInput ウィル・オ・ザ・ウィスプWebインプット
type WillOTheWispWebInput struct {
	BaseWebInput
	From *WillOTheWispWebZone `json:"from,omitempty"`
	To   *WillOTheWispWebZone `json:"to,omitempty"`
}

// WillOTheWispWebZone ゾーン指定
type WillOTheWispWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// WillOTheWispWebOutputTableauCard タブローカード出力
type WillOTheWispWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// WillOTheWispWebOutputHint ヒント出力
type WillOTheWispWebOutputHint struct {
	FromCol   int `json:"fromCol"`
	CardIndex int `json:"cardIndex"`
	ToCol     int `json:"toCol"`
}

// WillOTheWispWebOutput ウィル・オ・ザ・ウィスプWebアウトプット
// WillOTheWispWebOutputScoring はスコアの決まり (開始点 / 1手あたりの減点 / スート完成の加点)。
type WillOTheWispWebOutputScoring struct {
	Start       int `json:"start"`
	MovePenalty int `json:"movePenalty"`
	SuitBonus   int `json:"suitBonus"`
}

type WillOTheWispWebOutput struct {
	Tableau        [][]*WillOTheWispWebOutputTableauCard `json:"tableau"`
	StockCount     int                                   `json:"stockCount"`
	CompletedSuits int                                   `json:"completedSuits"`
	Score          int                                   `json:"score"`
	// Scoring はスコアの決まり (#5593)。数字が動く理由を説明するのに要る。
	// 訳文に焼き込むと、計算を変えたとき案内だけが古くなる。
	Scoring WillOTheWispWebOutputScoring `json:"scoring"`
	Hint    *WillOTheWispWebOutputHint   `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// WillOTheWispWebController ウィル・オ・ザ・ウィスプWebコントローラークラス
type WillOTheWispWebController = GameWebController[usecase.WillOTheWispInteractorIF, WillOTheWispWebInput, *WillOTheWispWebOutput]

// NewWillOTheWispWebController and NewWillOTheWispWebControllerWithProvider are
// the standard and provider-backed constructors for WillOTheWispWebController.
var NewWillOTheWispWebController, NewWillOTheWispWebControllerWithProvider = webControllerPair[usecase.WillOTheWispInteractorIF, WillOTheWispWebInput, *WillOTheWispWebOutput](
	newWillOTheWispDefaultOutput, willOTheWispDispatch,
)

func newWillOTheWispDefaultOutput(msg string) *WillOTheWispWebOutput {
	return &WillOTheWispWebOutput{
		Tableau:       make([][]*WillOTheWispWebOutputTableauCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func willOTheWispDispatch(bc *baseController, w http.ResponseWriter, si usecase.WillOTheWispInteractorIF, param WillOTheWispWebInput, newDefault func(string) *WillOTheWispWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.Reset())
	case "d", "deal":
		bc.writePresenterResponse(w, si.Deal())
	case "m", "move":
		return willOTheWispMoveDispatch(bc, w, si, param, newDefault)
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
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}

func willOTheWispMoveDispatch(bc *baseController, w http.ResponseWriter, si usecase.WillOTheWispInteractorIF, param WillOTheWispWebInput, newDefault func(string) *WillOTheWispWebOutput) bool {
	mv := tableauMove{haveFrom: param.From != nil, haveTo: param.To != nil}
	if param.From != nil {
		mv.fromZone, mv.fromCol, mv.fromCardIndex = param.From.Zone, param.From.Col, param.From.CardIndex
	}
	if param.To != nil {
		mv.toZone, mv.toCol = param.To.Zone, param.To.Col
	}
	return dispatchTableauOnlyMove(bc, w, mv, si.MoveTableauToTableau, newDefault)
}
