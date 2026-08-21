//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SalicLawWebInput サリカ法典 Web インプット
type SalicLawWebInput struct {
	BaseWebInput
	From *SalicLawWebZone `json:"from,omitempty"`
	To   *SalicLawWebZone `json:"to,omitempty"`
}

// SalicLawWebZone ゾーン指定。Zone は "tableau" / "foundation"。
// 捨て札は無く、めくった札は今の列に直接乗るので、waste / stock ゾーンは無い。
type SalicLawWebZone struct {
	Zone string `json:"zone"`
	// Col はタブロー列（0..7）。基礎札では不要。
	Col *int `json:"col,omitempty"`
}

// SalicLawWebOutputHint ヒント出力
type SalicLawWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// SalicLawWebOutput サリカ法典 Web アウトプット
type SalicLawWebOutput struct {
	Tableau    [][]*WebOutputCard `json:"tableau"`
	Foundation [][]*WebOutputCard `json:"foundation"`
	StockCount int                `json:"stockCount"`
	// Queens 場から抜いたクイーン 8 枚。飾りとして表示するだけで動かせない。
	Queens []*WebOutputCard `json:"queens"`
	// OpenPiles 土台の K が据わって使えるようになった列の数。
	OpenPiles int                    `json:"openPiles"`
	Hint      *SalicLawWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// SalicLawWebController サリカ法典 Web コントローラークラス
type SalicLawWebController = GameWebController[usecase.SalicLawInteractorIF, SalicLawWebInput, *SalicLawWebOutput]

// NewSalicLawWebController and NewSalicLawWebControllerWithProvider are the
// standard and provider-backed constructors for SalicLawWebController.
var NewSalicLawWebController, NewSalicLawWebControllerWithProvider = webControllerPair[usecase.SalicLawInteractorIF, SalicLawWebInput, *SalicLawWebOutput](
	newSalicLawDefaultOutput, salicLawDispatch,
)

func newSalicLawDefaultOutput(msg string) *SalicLawWebOutput {
	return &SalicLawWebOutput{
		Tableau:       make([][]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		Queens:        make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func salicLawDispatch(bc *baseController, w http.ResponseWriter, ci usecase.SalicLawInteractorIF, param SalicLawWebInput, newDefault func(string) *SalicLawWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ci.Draw())
	case "m", "move":
		return salicLawMoveDispatch(bc, w, ci, param, newDefault)
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

func salicLawMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.SalicLawInteractorIF, param SalicLawWebInput, newDefault func(string) *SalicLawWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToTableau(*param.From.Col, *param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
