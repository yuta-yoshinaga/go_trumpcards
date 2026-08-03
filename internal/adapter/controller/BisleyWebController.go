//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BisleyWebInput ビズリー Web インプット
type BisleyWebInput struct {
	BaseWebInput
	From *BisleyWebZone `json:"from,omitempty"`
	To   *BisleyWebZone `json:"to,omitempty"`
}

// BisleyWebZone ゾーン指定。Zone は "tableau" / "ace" / "king"。
type BisleyWebZone struct {
	Zone string `json:"zone"`
	Col  *int   `json:"col,omitempty"`
}

// BisleyWebOutputTableauCard タブローカード出力
type BisleyWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// BisleyWebOutputHint ヒント出力
type BisleyWebOutputHint struct {
	FromCol int    `json:"fromCol"`
	ToZone  string `json:"toZone"`
	ToIdx   int    `json:"toIdx"`
}

// BisleyWebOutput ビズリー Web アウトプット
type BisleyWebOutput struct {
	Tableau         [][]*BisleyWebOutputTableauCard `json:"tableau"`
	AceFoundations  [][]*WebOutputCard              `json:"aceFoundations"`
	KingFoundations [][]*WebOutputCard              `json:"kingFoundations"`
	Hint            *BisleyWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// BisleyWebController ビズリー Web コントローラークラス
type BisleyWebController = GameWebController[usecase.BisleyInteractorIF, BisleyWebInput, *BisleyWebOutput]

// NewBisleyWebController and NewBisleyWebControllerWithProvider are the standard
// and provider-backed constructors for BisleyWebController.
var NewBisleyWebController, NewBisleyWebControllerWithProvider = webControllerPair[usecase.BisleyInteractorIF, BisleyWebInput, *BisleyWebOutput](
	newBisleyDefaultOutput, bisleyDispatch,
)

func newBisleyDefaultOutput(msg string) *BisleyWebOutput {
	return &BisleyWebOutput{
		Tableau:         make([][]*BisleyWebOutputTableauCard, 0),
		AceFoundations:  make([][]*WebOutputCard, 0),
		KingFoundations: make([][]*WebOutputCard, 0),
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func bisleyDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BisleyInteractorIF, param BisleyWebInput, newDefault func(string) *BisleyWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return bisleyMoveDispatch(bc, w, bi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, bi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, bi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, bi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, bi.Reset, bi.Hint, bi.ActionLog)
	}
	return true
}

func bisleyMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BisleyInteractorIF, param BisleyWebInput, newDefault func(string) *BisleyWebOutput) bool {
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
		bc.writePresenterResponse(w, bi.MoveTableauToTableau(*param.From.Col, *param.To.Col))
	case "ace":
		bc.writePresenterResponse(w, bi.MoveTableauToAceFoundation(*param.From.Col))
	case "king":
		bc.writePresenterResponse(w, bi.MoveTableauToKingFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
