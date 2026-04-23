package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FortyThievesWebInput フォーティシーブスWebインプット
type FortyThievesWebInput struct {
	BaseWebInput
	From *FortyThievesWebZone `json:"from,omitempty"`
	To   *FortyThievesWebZone `json:"to,omitempty"`
}

// FortyThievesWebZone ゾーン指定
type FortyThievesWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// FortyThievesWebOutputTableauCard タブローカード出力
type FortyThievesWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// FortyThievesWebOutputHint ヒント出力
type FortyThievesWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// FortyThievesWebOutput フォーティシーブスWebアウトプット
type FortyThievesWebOutput struct {
	Tableau      [][]*FortyThievesWebOutputTableauCard `json:"tableau"`
	StockCount   int                                   `json:"stockCount"`
	Waste        []*WebOutputCard                      `json:"waste"`
	Foundation   [][]*WebOutputCard                    `json:"foundation"`
	Phase        int                                   `json:"phase"`
	MoveCount    int                                   `json:"moveCount"`
	CanUndo      bool                                  `json:"canUndo"`
	IsStalemate  bool                                  `json:"isStalemate"`
	UndoToEscape int                                   `json:"undoToEscape"`
	Hint         *FortyThievesWebOutputHint            `json:"hint,omitempty"`
	WebOutputBase
}

// FortyThievesWebController フォーティシーブスWebコントローラークラス
type FortyThievesWebController = GameWebController[usecase.FortyThievesInteractorIF, FortyThievesWebInput, *FortyThievesWebOutput]

// NewFortyThievesWebController and NewFortyThievesWebControllerWithProvider are
// the standard and provider-backed constructors for FortyThievesWebController.
var NewFortyThievesWebController, NewFortyThievesWebControllerWithProvider = webControllerPair[usecase.FortyThievesInteractorIF, FortyThievesWebInput, *FortyThievesWebOutput](
	newFortyThievesDefaultOutput, fortyThievesDispatch,
)

func newFortyThievesDefaultOutput(msg string) *FortyThievesWebOutput {
	return &FortyThievesWebOutput{
		Tableau:       make([][]*FortyThievesWebOutputTableauCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func fortyThievesDispatch(bc *baseController, w http.ResponseWriter, fi usecase.FortyThievesInteractorIF, param FortyThievesWebInput, newDefault func(string) *FortyThievesWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, fi.Draw())
	case "m", "move":
		return fortyThievesMoveDispatch(bc, w, fi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, fi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, fi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, fi.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, fi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, fi.Reset, fi.Hint, fi.ActionLog)
	}
	return true
}

func fortyThievesMoveDispatch(bc *baseController, w http.ResponseWriter, fi usecase.FortyThievesInteractorIF, param FortyThievesWebInput, newDefault func(string) *FortyThievesWebOutput) bool {
	if param.From == nil || param.To == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from and to are required."))
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "waste" && toZone == "tableau":
		if param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: to.col is required."))
			return true
		}
		bc.writePresenterResponse(w, fi.MoveWasteToTableau(*param.To.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, fi.MoveWasteToFoundation())
	case fromZone == "tableau" && toZone == "tableau":
		if param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col, from.cardIndex, to.col are required."))
			return true
		}
		bc.writePresenterResponse(w, fi.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if param.From.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col is required."))
			return true
		}
		bc.writePresenterResponse(w, fi.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
