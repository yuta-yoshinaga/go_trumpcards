package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RussianSolitaireWebInput ロシアンソリティアWebインプット
type RussianSolitaireWebInput struct {
	BaseWebInput
	From *RussianSolitaireWebZone `json:"from,omitempty"`
	To   *RussianSolitaireWebZone `json:"to,omitempty"`
}

// RussianSolitaireWebZone ゾーン指定
type RussianSolitaireWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// RussianSolitaireWebOutputHint ヒント出力
type RussianSolitaireWebOutputHint struct {
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// RussianSolitaireWebOutput ロシアンソリティアWebアウトプット
type RussianSolitaireWebOutput struct {
	Tableau    [][]*KlondikeWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                `json:"foundation"`
	Hint       *RussianSolitaireWebOutputHint    `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// RussianSolitaireWebController ロシアンソリティアWebコントローラークラス
type RussianSolitaireWebController = GameWebController[usecase.RussianSolitaireInteractorIF, RussianSolitaireWebInput, *RussianSolitaireWebOutput]

// NewRussianSolitaireWebController and NewRussianSolitaireWebControllerWithProvider are
// the standard and provider-backed constructors for RussianSolitaireWebController.
var NewRussianSolitaireWebController, NewRussianSolitaireWebControllerWithProvider = webControllerPair[usecase.RussianSolitaireInteractorIF, RussianSolitaireWebInput, *RussianSolitaireWebOutput](
	newRussianSolitaireDefaultOutput, russianSolitaireDispatch,
)

func newRussianSolitaireDefaultOutput(msg string) *RussianSolitaireWebOutput {
	return &RussianSolitaireWebOutput{
		Tableau:       make([][]*KlondikeWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func russianSolitaireDispatch(bc *baseController, w http.ResponseWriter, ri usecase.RussianSolitaireInteractorIF, param RussianSolitaireWebInput, newDefault func(string) *RussianSolitaireWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return russianSolitaireMoveDispatch(bc, w, ri, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ri.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ri.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ri.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, ri.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ri.Reset, ri.Hint, ri.ActionLog)
	}
	return true
}

func russianSolitaireMoveDispatch(bc *baseController, w http.ResponseWriter, ri usecase.RussianSolitaireInteractorIF, param RussianSolitaireWebInput, newDefault func(string) *RussianSolitaireWebOutput) bool {
	if param.From == nil || param.To == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from and to are required."))
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "tableau" && toZone == "tableau":
		if param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col, from.cardIndex, to.col are required."))
			return true
		}
		bc.writePresenterResponse(w, ri.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if param.From.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col is required."))
			return true
		}
		bc.writePresenterResponse(w, ri.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
