package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EightOffWebInput エイトオフWebインプット
type EightOffWebInput struct {
	BaseWebInput
	From *EightOffWebZone `json:"from,omitempty"`
	To   *EightOffWebZone `json:"to,omitempty"`
}

// EightOffWebZone ゾーン指定
type EightOffWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	Cell      *int   `json:"cell,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// EightOffWebOutputHint ヒント出力
type EightOffWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// EightOffWebOutput エイトオフWebアウトプット
type EightOffWebOutput struct {
	Tableau    [][]*WebOutputCard     `json:"tableau"`
	FreeCells  []*WebOutputCard       `json:"freeCells"`
	Foundation [][]*WebOutputCard     `json:"foundation"`
	Hint       *EightOffWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// EightOffWebController エイトオフWebコントローラークラス
type EightOffWebController = GameWebController[usecase.EightOffInteractorIF, EightOffWebInput, *EightOffWebOutput]

// NewEightOffWebController and NewEightOffWebControllerWithProvider are
// the standard and provider-backed constructors for EightOffWebController.
var NewEightOffWebController, NewEightOffWebControllerWithProvider = webControllerPair[usecase.EightOffInteractorIF, EightOffWebInput, *EightOffWebOutput](
	newEightOffDefaultOutput, eightOffDispatch,
)

func newEightOffDefaultOutput(msg string) *EightOffWebOutput {
	return &EightOffWebOutput{
		Tableau:       make([][]*WebOutputCard, 0),
		FreeCells:     make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func eightOffDispatch(bc *baseController, w http.ResponseWriter, ei usecase.EightOffInteractorIF, param EightOffWebInput, newDefault func(string) *EightOffWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return eightOffMoveDispatch(bc, w, ei, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ei.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ei.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ei.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, ei.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ei.Reset, ei.Hint, ei.ActionLog)
	}
	return true
}

func eightOffMoveDispatch(bc *baseController, w http.ResponseWriter, ei usecase.EightOffInteractorIF, param EightOffWebInput, newDefault func(string) *EightOffWebOutput) bool {
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
		bc.writePresenterResponse(w, ei.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if param.From.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col is required."))
			return true
		}
		bc.writePresenterResponse(w, ei.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "tableau" && toZone == "freecell":
		if param.From.Col == nil || param.To.Cell == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col and to.cell are required."))
			return true
		}
		bc.writePresenterResponse(w, ei.MoveTableauToFreeCell(*param.From.Col, *param.To.Cell))
	case fromZone == "freecell" && toZone == "tableau":
		if param.From.Cell == nil || param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.cell and to.col are required."))
			return true
		}
		bc.writePresenterResponse(w, ei.MoveFreeCellToTableau(*param.From.Cell, *param.To.Col))
	case fromZone == "freecell" && toZone == "foundation":
		if param.From.Cell == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.cell is required."))
			return true
		}
		bc.writePresenterResponse(w, ei.MoveFreeCellToFoundation(*param.From.Cell))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
