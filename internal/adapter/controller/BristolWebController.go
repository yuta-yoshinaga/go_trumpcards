//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BristolWebInput ブリストルWebインプット
type BristolWebInput struct {
	BaseWebInput
	From *BristolWebZone `json:"from,omitempty"`
	To   *BristolWebZone `json:"to,omitempty"`
}

// BristolWebZone ゾーン指定
type BristolWebZone struct {
	Zone string `json:"zone"`
	Col  *int   `json:"col,omitempty"`
}

// BristolWebOutputHint ヒント出力
type BristolWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromCol  int    `json:"fromCol"`
	ToZone   string `json:"toZone"`
	ToCol    int    `json:"toCol"`
}

// BristolWebOutput ブリストルWebアウトプット
type BristolWebOutput struct {
	Tableau    [][]*WebOutputCard `json:"tableau"`
	Fan        [][]*WebOutputCard `json:"fan"`
	StockCount int                `json:"stockCount"`
	Foundation [][]*WebOutputCard `json:"foundation"`
	Phase      int                `json:"phase"`
	MoveCount  int                `json:"moveCount"`
	CanUndo    bool               `json:"canUndo"`
	// IsStalemate は合法手が 1 つも無い状態か。ストックを作り直せないので
	// 普通に到達する (#5631)。
	IsStalemate bool `json:"isStalemate"`
	// UndoToEscape は膠着から抜けるのに必要なアンドゥ回数 (膠着でなければ 0)。
	UndoToEscape int                   `json:"undoToEscape"`
	Hint         *BristolWebOutputHint `json:"hint,omitempty"`
	// LegalTargets は移動元ごとの合法な移動先。キーは "tableau-0" / "fan-2"。
	// 選択中の札で実際に動かせる先だけを画面が示すために使う (#4813)。
	LegalTargets map[string]BristolWebOutputTargets `json:"legalTargets"`
	WebOutputBase
}

// BristolWebOutputTargets は 1 つの移動元から置ける先。
type BristolWebOutputTargets struct {
	Tableau    []int `json:"tableau"`
	Foundation []int `json:"foundation"`
}

// BristolWebController ブリストルWebコントローラークラス
type BristolWebController = GameWebController[usecase.BristolInteractorIF, BristolWebInput, *BristolWebOutput]

// NewBristolWebController and NewBristolWebControllerWithProvider are the
// standard and provider-backed constructors for BristolWebController.
var NewBristolWebController, NewBristolWebControllerWithProvider = webControllerPair[usecase.BristolInteractorIF, BristolWebInput, *BristolWebOutput](
	newBristolDefaultOutput, bristolDispatch,
)

func newBristolDefaultOutput(msg string) *BristolWebOutput {
	return &BristolWebOutput{
		Tableau:       make([][]*WebOutputCard, 0),
		Fan:           make([][]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func bristolDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BristolInteractorIF, param BristolWebInput, newDefault func(string) *BristolWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, bi.Draw())
	case "m", "move":
		return bristolMoveDispatch(bc, w, bi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, bi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, bi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, bi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, bi.Reset, bi.Hint, bi.ActionLog)
	}
	return true
}

func bristolMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BristolInteractorIF, param BristolWebInput, newDefault func(string) *BristolWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveTableauToTableau(*param.From.Col, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "fan" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveFanToTableau(*param.From.Col, *param.To.Col))
	case fromZone == "fan" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveFanToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
