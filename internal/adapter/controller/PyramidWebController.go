package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PyramidWebInput ピラミッドWebインプット
type PyramidWebInput struct {
	BaseWebInput
	Card1 *PyramidWebCard `json:"card1,omitempty"`
	Card2 *PyramidWebCard `json:"card2,omitempty"`
}

// PyramidWebCard カード位置指定
type PyramidWebCard struct {
	Zone string `json:"zone"` // "pyramid" or "waste"
	Row  *int   `json:"row,omitempty"`
	Col  *int   `json:"col,omitempty"`
}

// PyramidWebOutputCard ピラミッドカード出力
type PyramidWebOutputCard struct {
	Card    *WebOutputCard `json:"card"`
	Removed bool           `json:"removed"`
	Exposed bool           `json:"exposed"`
}

// PyramidWebOutputHint ヒント出力
type PyramidWebOutputHint struct {
	Type string `json:"type"`
	Row1 int    `json:"row1"`
	Col1 int    `json:"col1"`
	Row2 int    `json:"row2"`
	Col2 int    `json:"col2"`
}

// PyramidWebOutput ピラミッドWebアウトプット
type PyramidWebOutput struct {
	Pyramid     [][]*PyramidWebOutputCard `json:"pyramid"`
	StockCount  int                       `json:"stockCount"`
	Waste       []*WebOutputCard          `json:"waste"`
	Phase       int                       `json:"phase"`
	MoveCount   int                       `json:"moveCount"`
	CanUndo     bool                      `json:"canUndo"`
	IsStalemate bool                      `json:"isStalemate"`
	Hint        *PyramidWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
}

// PyramidWebController ピラミッドWebコントローラークラス
type PyramidWebController = GameWebController[usecase.PyramidInteractorIF, PyramidWebInput, *PyramidWebOutput]

// NewPyramidWebController コンストラクタ
func NewPyramidWebController(factory func() usecase.PyramidInteractorIF) *PyramidWebController {
	return NewGameWebController(factory, newPyramidDefaultOutput, pyramidDispatch)
}

// NewPyramidWebControllerWithProvider creates a PyramidWebController with an
// explicit SessionProvider (e.g. KV-backed for Workers).
func NewPyramidWebControllerWithProvider(
	provider SessionProvider[usecase.PyramidInteractorIF],
	factory func() usecase.PyramidInteractorIF,
) *PyramidWebController {
	return NewGameWebControllerWithProvider(provider, factory, newPyramidDefaultOutput, pyramidDispatch)
}

func newPyramidDefaultOutput(msg string) *PyramidWebOutput {
	return &PyramidWebOutput{
		Pyramid:       make([][]*PyramidWebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func pyramidDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PyramidInteractorIF, param PyramidWebInput, newDefault func(string) *PyramidWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.Reset())
	case "d", "draw":
		bc.writePresenterResponse(w, pi.Draw())
	case "rm", "remove":
		return pyramidRemoveDispatch(bc, w, pi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, pi.GiveUp())
	case "u", "undo":
		bc.writePresenterResponse(w, pi.Undo())
	default:
		return dispatchHintAndLog(param.Command, bc, w, pi.Hint, pi.ActionLog)
	}
	return true
}

func pyramidRemoveDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PyramidInteractorIF, param PyramidWebInput, newDefault func(string) *PyramidWebOutput) bool {
	if param.Card1 == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: card1 is required."))
		return true
	}

	c1 := param.Card1
	c2 := param.Card2

	switch {
	case c1.Zone == "pyramid" && c2 != nil && c2.Zone == "pyramid":
		// ピラミッド同士のペア除去
		if c1.Row == nil || c1.Col == nil || c2.Row == nil || c2.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: row and col are required for pyramid cards."))
			return true
		}
		bc.writePresenterResponse(w, pi.RemovePair(*c1.Row, *c1.Col, *c2.Row, *c2.Col))
	case c1.Zone == "pyramid" && c2 == nil:
		// ピラミッドのキング除去
		if c1.Row == nil || c1.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: row and col are required."))
			return true
		}
		bc.writePresenterResponse(w, pi.RemoveKing(*c1.Row, *c1.Col))
	case c1.Zone == "waste" && c2 != nil && c2.Zone == "pyramid":
		// ウェイスト+ピラミッドのペア除去
		if c2.Row == nil || c2.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: row and col are required for pyramid card."))
			return true
		}
		bc.writePresenterResponse(w, pi.RemoveWithWaste(*c2.Row, *c2.Col))
	case c1.Zone == "pyramid" && c2 != nil && c2.Zone == "waste":
		// ピラミッド+ウェイストのペア除去 (逆順もサポート)
		if c1.Row == nil || c1.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: row and col are required for pyramid card."))
			return true
		}
		bc.writePresenterResponse(w, pi.RemoveWithWaste(*c1.Row, *c1.Col))
	case c1.Zone == "waste" && c2 == nil:
		// ウェイストのキング除去
		bc.writePresenterResponse(w, pi.RemoveWasteKing())
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid remove combination."))
	}
	return true
}
