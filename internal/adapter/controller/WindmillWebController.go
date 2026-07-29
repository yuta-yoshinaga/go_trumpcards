//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WindmillWebInput ウィンドミル Web インプット
type WindmillWebInput struct {
	BaseWebInput
	From *WindmillWebZone `json:"from,omitempty"`
	To   *WindmillWebZone `json:"to,omitempty"`
}

// WindmillWebZone ゾーン指定。Zone は "sail" / "waste" / "corner" / "center"。
type WindmillWebZone struct {
	Zone string `json:"zone"`
	// Col は帆（0..7）または四隅（0..3）。捨て札と中央では不要。
	Col *int `json:"col,omitempty"`
}

// WindmillWebOutputHint ヒント出力
type WindmillWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// WindmillWebOutput ウィンドミル Web アウトプット
type WindmillWebOutput struct {
	// Sails は 8 枠固定。補充できなくなった枠は null になる。
	Sails      []*WebOutputCard   `json:"sails"`
	Center     []*WebOutputCard   `json:"center"`
	Corners    [][]*WebOutputCard `json:"corners"`
	StockCount int                `json:"stockCount"`
	Waste      []*WebOutputCard   `json:"waste"`
	// TransferBlocked が真の間、四隅→中央の引き戻しは拒否される。
	TransferBlocked bool                   `json:"transferBlocked"`
	Hint            *WindmillWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// WindmillWebController ウィンドミル Web コントローラークラス
type WindmillWebController = GameWebController[usecase.WindmillInteractorIF, WindmillWebInput, *WindmillWebOutput]

// NewWindmillWebController and NewWindmillWebControllerWithProvider are the
// standard and provider-backed constructors for WindmillWebController.
var NewWindmillWebController, NewWindmillWebControllerWithProvider = webControllerPair[usecase.WindmillInteractorIF, WindmillWebInput, *WindmillWebOutput](
	newWindmillDefaultOutput, windmillDispatch,
)

func newWindmillDefaultOutput(msg string) *WindmillWebOutput {
	return &WindmillWebOutput{
		Sails:         make([]*WebOutputCard, 0),
		Center:        make([]*WebOutputCard, 0),
		Corners:       make([][]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func windmillDispatch(bc *baseController, w http.ResponseWriter, wi usecase.WindmillInteractorIF, param WindmillWebInput, newDefault func(string) *WindmillWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, wi.Draw())
	case "m", "move":
		return windmillMoveDispatch(bc, w, wi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, wi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, wi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, wi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, wi.Reset, wi.Hint, wi.ActionLog)
	}
	return true
}

func windmillMoveDispatch(bc *baseController, w http.ResponseWriter, wi usecase.WindmillInteractorIF, param WindmillWebInput, newDefault func(string) *WindmillWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "sail" && toZone == "center":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.MoveSailToCenter(*param.From.Col))
	case fromZone == "sail" && toZone == "corner":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.MoveSailToCorner(*param.From.Col, *param.To.Col))
	case fromZone == "waste" && toZone == "center":
		bc.writePresenterResponse(w, wi.MoveWasteToCenter())
	case fromZone == "waste" && toZone == "corner":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.MoveWasteToCorner(*param.To.Col))
	case fromZone == "corner" && toZone == "center":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.MoveCornerToCenter(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
