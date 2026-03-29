package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TriPeaksWebInput トリピークスWebインプット
type TriPeaksWebInput struct {
	BaseWebInput
	Row *int `json:"row,omitempty"`
	Col *int `json:"col,omitempty"`
}

// TriPeaksWebOutputCard タブローカード出力
type TriPeaksWebOutputCard struct {
	Card    *WebOutputCard `json:"card"`
	Removed bool           `json:"removed"`
	Exposed bool           `json:"exposed"`
}

// TriPeaksWebOutputHint ヒント出力
type TriPeaksWebOutputHint struct {
	Type string `json:"type"`
	Row  int    `json:"row"`
	Col  int    `json:"col"`
}

// TriPeaksWebOutput トリピークスWebアウトプット
type TriPeaksWebOutput struct {
	Layout      [][]*TriPeaksWebOutputCard `json:"layout"`
	StockCount  int                        `json:"stockCount"`
	Waste       []*WebOutputCard           `json:"waste"`
	Phase       int                        `json:"phase"`
	MoveCount   int                        `json:"moveCount"`
	CanUndo     bool                       `json:"canUndo"`
	IsStalemate bool                       `json:"isStalemate"`
	Hint        *TriPeaksWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
}

// TriPeaksWebController トリピークスWebコントローラークラス
type TriPeaksWebController = GameWebController[usecase.TriPeaksInteractorIF, TriPeaksWebInput, *TriPeaksWebOutput]

// NewTriPeaksWebController コンストラクタ
func NewTriPeaksWebController(factory func() usecase.TriPeaksInteractorIF) *TriPeaksWebController {
	return NewGameWebController(factory, newTriPeaksDefaultOutput, triPeaksDispatch)
}

// NewTriPeaksWebControllerWithProvider creates a TriPeaksWebController with an
// explicit SessionProvider (e.g. KV-backed for Workers).
func NewTriPeaksWebControllerWithProvider(
	provider SessionProvider[usecase.TriPeaksInteractorIF],
	factory func() usecase.TriPeaksInteractorIF,
) *TriPeaksWebController {
	return NewGameWebControllerWithProvider(provider, factory, newTriPeaksDefaultOutput, triPeaksDispatch)
}

func newTriPeaksDefaultOutput(msg string) *TriPeaksWebOutput {
	return &TriPeaksWebOutput{
		Layout:        make([][]*TriPeaksWebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func triPeaksDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TriPeaksInteractorIF, param TriPeaksWebInput, newDefault func(string) *TriPeaksWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.Reset())
	case "d", "draw":
		bc.writePresenterResponse(w, ti.Draw())
	case "rm", "remove":
		if param.Row == nil || param.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: row and col are required."))
			return true
		}
		bc.writePresenterResponse(w, ti.Remove(*param.Row, *param.Col))
	case "g", "giveup":
		bc.writePresenterResponse(w, ti.GiveUp())
	case "h", "hint":
		bc.writePresenterResponse(w, ti.Hint())
	case "log", "l":
		bc.writePresenterResponse(w, ti.ActionLog())
	case "u", "undo":
		bc.writePresenterResponse(w, ti.Undo())
	default:
		return false
	}
	return true
}
