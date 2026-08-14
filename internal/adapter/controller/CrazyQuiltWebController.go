//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CrazyQuiltWebInput クレイジーキルト Web インプット
type CrazyQuiltWebInput struct {
	BaseWebInput
	From *CrazyQuiltWebZone `json:"from,omitempty"`
	To   *CrazyQuiltWebZone `json:"to,omitempty"`
}

// CrazyQuiltWebZone ゾーン指定。Zone は "quilt" / "waste" / "foundation"。
type CrazyQuiltWebZone struct {
	Zone string `json:"zone"`
	// Col はキルトのマス番号（0..63、row*8+col）。捨て札・基礎札では不要。
	Col *int `json:"col,omitempty"`
}

// CrazyQuiltWebOutputHint ヒント出力
type CrazyQuiltWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// CrazyQuiltWebOutput クレイジーキルト Web アウトプット
type CrazyQuiltWebOutput struct {
	// Quilt はマス番号順（row*8+col）。取り除いたマスは null。
	Quilt []*WebOutputCard `json:"quilt"`
	// Available はマスごとの可動判定。短辺が露出しているかは向きに依存するので
	// サーバが計算して送る（フロントで再実装すると規則が 2 か所に散る）。
	Available []bool `json:"available"`
	// FoundationAscending は基礎札ごとの向き。true が A→K、false が K→A。
	FoundationAscending []bool `json:"foundationAscending"`
	// RedealsLeft は残りの組み直し回数。
	RedealsLeft int                      `json:"redealsLeft"`
	Foundation  [][]*WebOutputCard       `json:"foundation"`
	StockCount  int                      `json:"stockCount"`
	Waste       []*WebOutputCard         `json:"waste"`
	Hint        *CrazyQuiltWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// CrazyQuiltWebController クレイジーキルト Web コントローラークラス
type CrazyQuiltWebController = GameWebController[usecase.CrazyQuiltInteractorIF, CrazyQuiltWebInput, *CrazyQuiltWebOutput]

// NewCrazyQuiltWebController and NewCrazyQuiltWebControllerWithProvider are the
// standard and provider-backed constructors for CrazyQuiltWebController.
var NewCrazyQuiltWebController, NewCrazyQuiltWebControllerWithProvider = webControllerPair[usecase.CrazyQuiltInteractorIF, CrazyQuiltWebInput, *CrazyQuiltWebOutput](
	newCrazyQuiltDefaultOutput, crazyquiltDispatch,
)

func newCrazyQuiltDefaultOutput(msg string) *CrazyQuiltWebOutput {
	return &CrazyQuiltWebOutput{
		Quilt:               make([]*WebOutputCard, 0),
		Available:           make([]bool, 0),
		FoundationAscending: make([]bool, 0),
		Foundation:          make([][]*WebOutputCard, 0),
		Waste:               make([]*WebOutputCard, 0),
		WebOutputBase:       WebOutputBase{Message: msg},
	}
}

func crazyquiltDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CrazyQuiltInteractorIF, param CrazyQuiltWebInput, newDefault func(string) *CrazyQuiltWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "m", "move":
		return crazyquiltMoveDispatch(bc, w, ci, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ci.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ci.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ci.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ci.Reset, ci.Hint, ci.ActionLog)
	}
	return true
}

func crazyquiltMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CrazyQuiltInteractorIF, param CrazyQuiltWebInput, newDefault func(string) *CrazyQuiltWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "quilt" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveQuiltToFoundation(*param.From.Col))
	case fromZone == "quilt" && toZone == "waste":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveQuiltToWaste(*param.From.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, ci.MoveWasteToFoundation())
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
