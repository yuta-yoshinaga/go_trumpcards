package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SeahavenTowersWebInput シーヘイブンタワーズWebインプット
type SeahavenTowersWebInput struct {
	BaseWebInput
	From *SeahavenTowersWebZone `json:"from,omitempty"`
	To   *SeahavenTowersWebZone `json:"to,omitempty"`
}

// SeahavenTowersWebZone ゾーン指定
type SeahavenTowersWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	Cell      *int   `json:"cell,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// SeahavenTowersWebOutputHint ヒント出力
type SeahavenTowersWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// SeahavenTowersWebOutput シーヘイブンタワーズWebアウトプット
type SeahavenTowersWebOutput struct {
	Tableau       [][]*WebOutputCard           `json:"tableau"`
	ReservedCells []*WebOutputCard             `json:"reservedCells"`
	Foundation    [][]*WebOutputCard           `json:"foundation"`
	Hint          *SeahavenTowersWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// SeahavenTowersWebController シーヘイブンタワーズWebコントローラークラス
type SeahavenTowersWebController = GameWebController[usecase.SeahavenTowersInteractorIF, SeahavenTowersWebInput, *SeahavenTowersWebOutput]

// NewSeahavenTowersWebController and NewSeahavenTowersWebControllerWithProvider are
// the standard and provider-backed constructors for SeahavenTowersWebController.
var NewSeahavenTowersWebController, NewSeahavenTowersWebControllerWithProvider = webControllerPair[usecase.SeahavenTowersInteractorIF, SeahavenTowersWebInput, *SeahavenTowersWebOutput](
	newSeahavenTowersDefaultOutput, seahavenTowersDispatch,
)

func newSeahavenTowersDefaultOutput(msg string) *SeahavenTowersWebOutput {
	return &SeahavenTowersWebOutput{
		Tableau:       make([][]*WebOutputCard, 0),
		ReservedCells: make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func seahavenTowersDispatch(bc *baseController, w http.ResponseWriter, si usecase.SeahavenTowersInteractorIF, param SeahavenTowersWebInput, newDefault func(string) *SeahavenTowersWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return seahavenTowersMoveDispatch(bc, w, si, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, si.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, si.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, si.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, si.Reset, si.Hint, si.ActionLog)
	}
	return true
}

func seahavenTowersMoveDispatch(bc *baseController, w http.ResponseWriter, si usecase.SeahavenTowersInteractorIF, param SeahavenTowersWebInput, newDefault func(string) *SeahavenTowersWebOutput) bool {
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
		bc.writePresenterResponse(w, si.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if param.From.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col is required."))
			return true
		}
		bc.writePresenterResponse(w, si.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "tableau" && toZone == "reserved":
		if param.From.Col == nil || param.To.Cell == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col and to.cell are required."))
			return true
		}
		bc.writePresenterResponse(w, si.MoveTableauToFreeCell(*param.From.Col, *param.To.Cell))
	case fromZone == "reserved" && toZone == "tableau":
		if param.From.Cell == nil || param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.cell and to.col are required."))
			return true
		}
		bc.writePresenterResponse(w, si.MoveFreeCellToTableau(*param.From.Cell, *param.To.Col))
	case fromZone == "reserved" && toZone == "foundation":
		if param.From.Cell == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.cell is required."))
			return true
		}
		bc.writePresenterResponse(w, si.MoveFreeCellToFoundation(*param.From.Cell))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
