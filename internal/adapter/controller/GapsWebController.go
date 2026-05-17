package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GapsWebInput はGapsゲームのWebリクエスト入力。
type GapsWebInput struct {
	BaseWebInput
	From *GapsWebZone `json:"from,omitempty"`
	To   *GapsWebZone `json:"to,omitempty"`
}

// GapsWebZone はGapsの座標指定。
type GapsWebZone struct {
	Zone string `json:"zone"` // 常に "grid"
	Row  *int   `json:"row,omitempty"`
	Col  *int   `json:"col,omitempty"`
}

// GapsWebOutputHint はGapsのヒント出力。
type GapsWebOutputHint struct {
	FromRow int `json:"fromRow"`
	FromCol int `json:"fromCol"`
	ToRow   int `json:"toRow"`
	ToCol   int `json:"toCol"`
}

// GapsWebOutput はGapsゲームのWebレスポンス出力。
type GapsWebOutput struct {
	Grid             [][]*WebOutputCard `json:"grid"`
	RedealsUsed      int                `json:"redealsUsed"`
	RedealsRemaining int                `json:"redealsRemaining"`
	Hint             *GapsWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// GapsWebController はGapsゲームのWebコントローラー型。
type GapsWebController = GameWebController[usecase.GapsInteractorIF, GapsWebInput, *GapsWebOutput]

// NewGapsWebController と NewGapsWebControllerWithProvider はGapsWebControllerの
// 標準コンストラクタとプロバイダー版コンストラクタ。
var NewGapsWebController, NewGapsWebControllerWithProvider = webControllerPair[usecase.GapsInteractorIF, GapsWebInput, *GapsWebOutput](
	newGapsDefaultOutput, gapsDispatch,
)

func newGapsDefaultOutput(msg string) *GapsWebOutput {
	return &GapsWebOutput{
		Grid:          make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func gapsDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GapsInteractorIF, param GapsWebInput, newDefault func(string) *GapsWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return gapsMoveDispatch(bc, w, gi, param, newDefault)
	case "rd", "redeal":
		bc.writePresenterResponse(w, gi.Redeal())
	case "g", "giveup":
		bc.writePresenterResponse(w, gi.GiveUp())
	case "u", "undo":
		bc.writePresenterResponse(w, gi.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, gi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, gi.Reset, gi.Hint, gi.ActionLog)
	}
	return true
}

func gapsMoveDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GapsInteractorIF, param GapsWebInput, newDefault func(string) *GapsWebOutput) bool {
	if param.From == nil || param.To == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from and to are required."))
		return true
	}
	if param.From.Row == nil || param.From.Col == nil || param.To.Row == nil || param.To.Col == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: row and col are required."))
		return true
	}
	bc.writePresenterResponse(w, gi.Move(*param.From.Row, *param.From.Col, *param.To.Row, *param.To.Col))
	return true
}
