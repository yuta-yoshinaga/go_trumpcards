//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DuchessWebInput ダッチェス Web インプット
type DuchessWebInput struct {
	BaseWebInput
	From *DuchessWebZone `json:"from,omitempty"`
	To   *DuchessWebZone `json:"to,omitempty"`
}

// DuchessWebZone ゾーン指定。Zone は "reserve" / "waste" / "tableau" / "foundation"。
type DuchessWebZone struct {
	Zone string `json:"zone"`
	// Col はリザーブ扇（0..3）またはタブロー列（0..3）。ウェイストと基礎札では不要。
	Col *int `json:"col,omitempty"`
	// CardIndex は移動する連番グループの先頭。省略すると最上段 1 枚だけを動かす。
	CardIndex *int `json:"cardIndex,omitempty"`
}

// DuchessWebOutputTableauCard タブローカード出力
type DuchessWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// DuchessWebOutputHint ヒント出力
type DuchessWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromIdx   int    `json:"fromIdx"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToIdx     int    `json:"toIdx"`
}

// DuchessWebOutput ダッチェス Web アウトプット
type DuchessWebOutput struct {
	Reserve    [][]*WebOutputCard               `json:"reserve"`
	Tableau    [][]*DuchessWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard               `json:"foundation"`
	StockCount int                              `json:"stockCount"`
	Waste      []*WebOutputCard                 `json:"waste"`
	// BaseRank は 4 つの基礎札が始まるランク。0 は「まだ選ばれていない」。
	BaseRank int `json:"baseRank"`
	// AwaitingBaseRank が真の間は、リザーブから 1 枚選ぶ以外に何もできない。
	AwaitingBaseRank bool                  `json:"awaitingBaseRank"`
	Hint             *DuchessWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// DuchessWebController ダッチェス Web コントローラークラス
type DuchessWebController = GameWebController[usecase.DuchessInteractorIF, DuchessWebInput, *DuchessWebOutput]

// NewDuchessWebController and NewDuchessWebControllerWithProvider are the
// standard and provider-backed constructors for DuchessWebController.
var NewDuchessWebController, NewDuchessWebControllerWithProvider = webControllerPair[usecase.DuchessInteractorIF, DuchessWebInput, *DuchessWebOutput](
	newDuchessDefaultOutput, duchessDispatch,
)

func newDuchessDefaultOutput(msg string) *DuchessWebOutput {
	return &DuchessWebOutput{
		Reserve:       make([][]*WebOutputCard, 0),
		Tableau:       make([][]*DuchessWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func duchessDispatch(bc *baseController, w http.ResponseWriter, di usecase.DuchessInteractorIF, param DuchessWebInput, newDefault func(string) *DuchessWebOutput) bool {
	switch param.Command {
	case "base", "choosebase":
		if !requireParam(bc, w, newDefault, param.From == nil || param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.ChooseBaseRank(*param.From.Col))
	case "d", "draw":
		bc.writePresenterResponse(w, di.Draw())
	case "m", "move":
		return duchessMoveDispatch(bc, w, di, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, di.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, di.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, di.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, di.Reset, di.Hint, di.ActionLog)
	}
	return true
}

func duchessMoveDispatch(bc *baseController, w http.ResponseWriter, di usecase.DuchessInteractorIF, param DuchessWebInput, newDefault func(string) *DuchessWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "reserve" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.MoveReserveToFoundation(*param.From.Col))
	case fromZone == "reserve" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, di.MoveReserveToTableau(*param.From.Col, *param.To.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, di.MoveWasteToFoundation())
	case fromZone == "waste" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.MoveWasteToTableau(*param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		// cardIndex は連番グループの先頭。省略時は -1 = 最上段 1 枚。
		cardIndex := -1
		if param.From.CardIndex != nil {
			cardIndex = *param.From.CardIndex
		}
		bc.writePresenterResponse(w, di.MoveTableauToTableau(*param.From.Col, cardIndex, *param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
