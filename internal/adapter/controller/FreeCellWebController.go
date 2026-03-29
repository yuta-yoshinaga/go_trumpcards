package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FreeCellWebInput フリーセルWebインプット
type FreeCellWebInput struct {
	BaseWebInput
	From *FreeCellWebZone `json:"from,omitempty"`
	To   *FreeCellWebZone `json:"to,omitempty"`
}

// FreeCellWebZone ゾーン指定
type FreeCellWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	Cell      *int   `json:"cell,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// FreeCellWebOutputHint ヒント出力
type FreeCellWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// FreeCellWebOutput フリーセルWebアウトプット
type FreeCellWebOutput struct {
	Tableau     [][]*WebOutputCard     `json:"tableau"`
	FreeCells   []*WebOutputCard       `json:"freeCells"`
	Foundation  [][]*WebOutputCard     `json:"foundation"`
	Phase       int                    `json:"phase"`
	MoveCount   int                    `json:"moveCount"`
	CanUndo     bool                   `json:"canUndo"`
	IsStalemate bool                   `json:"isStalemate"`
	Hint        *FreeCellWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// FreeCellWebController フリーセルWebコントローラークラス
type FreeCellWebController = GameWebController[usecase.FreeCellInteractorIF, FreeCellWebInput, *FreeCellWebOutput]

// NewFreeCellWebController コンストラクタ
func NewFreeCellWebController(factory func() usecase.FreeCellInteractorIF) *FreeCellWebController {
	return NewGameWebController(factory, newFreeCellDefaultOutput, freeCellDispatch)
}

// NewFreeCellWebControllerWithProvider creates a FreeCellWebController with an
// explicit SessionProvider (e.g. KV-backed for Workers).
func NewFreeCellWebControllerWithProvider(
	provider SessionProvider[usecase.FreeCellInteractorIF],
	factory func() usecase.FreeCellInteractorIF,
) *FreeCellWebController {
	return NewGameWebControllerWithProvider(provider, factory, newFreeCellDefaultOutput, freeCellDispatch)
}

func newFreeCellDefaultOutput(msg string) *FreeCellWebOutput {
	return &FreeCellWebOutput{
		Tableau:       make([][]*WebOutputCard, 0),
		FreeCells:     make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func freeCellDispatch(bc *baseController, w http.ResponseWriter, fi usecase.FreeCellInteractorIF, param FreeCellWebInput, newDefault func(string) *FreeCellWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, fi.Reset())
	case "m", "move":
		return freeCellMoveDispatch(bc, w, fi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, fi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, fi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, fi.Undo())
	default:
		return dispatchHintAndLog(param.Command, bc, w, fi.Hint, fi.ActionLog)
	}
	return true
}

func freeCellMoveDispatch(bc *baseController, w http.ResponseWriter, fi usecase.FreeCellInteractorIF, param FreeCellWebInput, newDefault func(string) *FreeCellWebOutput) bool {
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
		bc.writePresenterResponse(w, fi.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if param.From.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col is required."))
			return true
		}
		bc.writePresenterResponse(w, fi.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "tableau" && toZone == "freecell":
		if param.From.Col == nil || param.To.Cell == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col and to.cell are required."))
			return true
		}
		bc.writePresenterResponse(w, fi.MoveTableauToFreeCell(*param.From.Col, *param.To.Cell))
	case fromZone == "freecell" && toZone == "tableau":
		if param.From.Cell == nil || param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.cell and to.col are required."))
			return true
		}
		bc.writePresenterResponse(w, fi.MoveFreeCellToTableau(*param.From.Cell, *param.To.Col))
	case fromZone == "freecell" && toZone == "foundation":
		if param.From.Cell == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.cell is required."))
			return true
		}
		bc.writePresenterResponse(w, fi.MoveFreeCellToFoundation(*param.From.Cell))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
