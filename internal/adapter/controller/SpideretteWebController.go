//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpideretteWebInput スパイダレットWebインプット
type SpideretteWebInput struct {
	BaseWebInput
	From *SpideretteWebZone `json:"from,omitempty"`
	To   *SpideretteWebZone `json:"to,omitempty"`
}

// SpideretteWebZone ゾーン指定
type SpideretteWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// SpideretteWebOutputTableauCard タブローカード出力
type SpideretteWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// SpideretteWebOutputHint ヒント出力
type SpideretteWebOutputHint struct {
	FromCol   int `json:"fromCol"`
	CardIndex int `json:"cardIndex"`
	ToCol     int `json:"toCol"`
}

// SpideretteWebOutput スパイダレットWebアウトプット
// SpideretteWebOutputScoring はスコアの決まり (開始点 / 1手あたりの減点 / スート完成の加点)。
type SpideretteWebOutputScoring struct {
	Start       int `json:"start"`
	MovePenalty int `json:"movePenalty"`
	SuitBonus   int `json:"suitBonus"`
}

type SpideretteWebOutput struct {
	Tableau        [][]*SpideretteWebOutputTableauCard `json:"tableau"`
	StockCount     int                                 `json:"stockCount"`
	CompletedSuits int                                 `json:"completedSuits"`
	Score          int                                 `json:"score"`
	// Scoring はスコアの決まり (#5593)。数字が動く理由を説明するのに要る。
	// 訳文に焼き込むと、計算を変えたとき案内だけが古くなる。
	Scoring SpideretteWebOutputScoring `json:"scoring"`
	Hint    *SpideretteWebOutputHint   `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// SpideretteWebController スパイダレットWebコントローラークラス
type SpideretteWebController = GameWebController[usecase.SpideretteInteractorIF, SpideretteWebInput, *SpideretteWebOutput]

// NewSpideretteWebController and NewSpideretteWebControllerWithProvider are
// the standard and provider-backed constructors for SpideretteWebController.
var NewSpideretteWebController, NewSpideretteWebControllerWithProvider = webControllerPair[usecase.SpideretteInteractorIF, SpideretteWebInput, *SpideretteWebOutput](
	newSpideretteDefaultOutput, spideretteDispatch,
)

func newSpideretteDefaultOutput(msg string) *SpideretteWebOutput {
	return &SpideretteWebOutput{
		Tableau:       make([][]*SpideretteWebOutputTableauCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func spideretteDispatch(bc *baseController, w http.ResponseWriter, si usecase.SpideretteInteractorIF, param SpideretteWebInput, newDefault func(string) *SpideretteWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.Reset())
	case "d", "deal":
		bc.writePresenterResponse(w, si.Deal())
	case "m", "move":
		return spideretteMoveDispatch(bc, w, si, param, newDefault)
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

func spideretteMoveDispatch(bc *baseController, w http.ResponseWriter, si usecase.SpideretteInteractorIF, param SpideretteWebInput, newDefault func(string) *SpideretteWebOutput) bool {
	mv := tableauMove{haveFrom: param.From != nil, haveTo: param.To != nil}
	if param.From != nil {
		mv.fromZone, mv.fromCol, mv.fromCardIndex = param.From.Zone, param.From.Col, param.From.CardIndex
	}
	if param.To != nil {
		mv.toZone, mv.toCol = param.To.Zone, param.To.Col
	}
	return dispatchTableauOnlyMove(bc, w, mv, si.MoveTableauToTableau, newDefault)
}
