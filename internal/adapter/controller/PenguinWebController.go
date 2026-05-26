package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PenguinWebInput ペンギンWebインプット
type PenguinWebInput struct {
	BaseWebInput
	From *PenguinWebZone `json:"from,omitempty"`
	To   *PenguinWebZone `json:"to,omitempty"`
}

// PenguinWebZone ゾーン指定
type PenguinWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	Cell      *int   `json:"cell,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// PenguinWebOutputHint ヒント出力
type PenguinWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// PenguinWebOutput ペンギンWebアウトプット
type PenguinWebOutput struct {
	Tableau    [][]*WebOutputCard    `json:"tableau"`
	FreeCells  []*WebOutputCard      `json:"freeCells"`
	Foundation [][]*WebOutputCard    `json:"foundation"`
	BaseRank   int                   `json:"baseRank"`
	Hint       *PenguinWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// PenguinWebController ペンギンWebコントローラークラス
type PenguinWebController = GameWebController[usecase.PenguinInteractorIF, PenguinWebInput, *PenguinWebOutput]

// NewPenguinWebController and NewPenguinWebControllerWithProvider are
// the standard and provider-backed constructors for PenguinWebController.
var NewPenguinWebController, NewPenguinWebControllerWithProvider = webControllerPair[usecase.PenguinInteractorIF, PenguinWebInput, *PenguinWebOutput](
	newPenguinDefaultOutput, penguinDispatch,
)

func newPenguinDefaultOutput(msg string) *PenguinWebOutput {
	return &PenguinWebOutput{
		Tableau:       make([][]*WebOutputCard, 0),
		FreeCells:     make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func penguinDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PenguinInteractorIF, param PenguinWebInput, newDefault func(string) *PenguinWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return penguinMoveDispatch(bc, w, pi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, pi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, pi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, pi.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, pi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, pi.Reset, pi.Hint, pi.ActionLog)
	}
	return true
}

func penguinMoveDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PenguinInteractorIF, param PenguinWebInput, newDefault func(string) *PenguinWebOutput) bool {
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
		bc.writePresenterResponse(w, pi.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if param.From.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col is required."))
			return true
		}
		bc.writePresenterResponse(w, pi.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "tableau" && toZone == "freecell":
		if param.From.Col == nil || param.To.Cell == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col and to.cell are required."))
			return true
		}
		bc.writePresenterResponse(w, pi.MoveTableauToFreeCell(*param.From.Col, *param.To.Cell))
	case fromZone == "freecell" && toZone == "tableau":
		if param.From.Cell == nil || param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.cell and to.col are required."))
			return true
		}
		bc.writePresenterResponse(w, pi.MoveFreeCellToTableau(*param.From.Cell, *param.To.Col))
	case fromZone == "freecell" && toZone == "foundation":
		if param.From.Cell == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.cell is required."))
			return true
		}
		bc.writePresenterResponse(w, pi.MoveFreeCellToFoundation(*param.From.Cell))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
