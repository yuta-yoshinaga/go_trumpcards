//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BraidWebInput ブレイド Web インプット
type BraidWebInput struct {
	BaseWebInput
	From *BraidWebZone `json:"from,omitempty"`
	To   *BraidWebZone `json:"to,omitempty"`
	// Ascending は "dir" コマンドで積む向きを指定する。
	Ascending *bool `json:"ascending,omitempty"`
}

// BraidWebZone ゾーン指定。Zone は "braid" / "field" / "helper" / "waste" / "foundation"。
type BraidWebZone struct {
	Zone string `json:"zone"`
	// Col はブレイド札（0..3）またはヘルパー（0..7）の枠番号。
	// ブレイド・捨て札・基礎札では不要。
	Col *int `json:"col,omitempty"`
}

// BraidWebOutputHint ヒント出力
type BraidWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// BraidWebOutput ブレイド Web アウトプット
type BraidWebOutput struct {
	Braid []*WebOutputCard `json:"braid"`
	// Fields / Helpers は空き枠を null のまま送る。詰めるとインデックスがずれ、
	// ヒントの枠番号が画面と食い違う。
	Fields     []*WebOutputCard   `json:"fields"`
	Helpers    []*WebOutputCard   `json:"helpers"`
	Foundation [][]*WebOutputCard `json:"foundation"`
	StockCount int                `json:"stockCount"`
	Waste      []*WebOutputCard   `json:"waste"`
	// BaseRank は 8 つの基礎札すべての開始ランク。配り切った時点で決まる。
	BaseRank int `json:"baseRank"`
	// Direction は 0=未選択 / 1=昇順 / 2=降順。
	Direction int `json:"direction"`
	// AwaitingDirection が真の間は、向きを選ぶまで基礎札に触れられない。
	AwaitingDirection bool `json:"awaitingDirection"`
	// RedealsLeft は残りのめくり直し回数（初期値 2）。
	RedealsLeft int                 `json:"redealsLeft"`
	CanRedeal   bool                `json:"canRedeal"`
	Hint        *BraidWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// BraidWebController ブレイド Web コントローラークラス
type BraidWebController = GameWebController[usecase.BraidInteractorIF, BraidWebInput, *BraidWebOutput]

// NewBraidWebController and NewBraidWebControllerWithProvider are the
// standard and provider-backed constructors for BraidWebController.
var NewBraidWebController, NewBraidWebControllerWithProvider = webControllerPair[usecase.BraidInteractorIF, BraidWebInput, *BraidWebOutput](
	newBraidDefaultOutput, braidDispatch,
)

func newBraidDefaultOutput(msg string) *BraidWebOutput {
	return &BraidWebOutput{
		Braid:         make([]*WebOutputCard, 0),
		Fields:        make([]*WebOutputCard, 0),
		Helpers:       make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func braidDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BraidInteractorIF, param BraidWebInput, newDefault func(string) *BraidWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, bi.Draw())
	case "dir", "direction":
		if !requireParam(bc, w, newDefault, param.Ascending == nil, "param error: ascending is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.ChooseDirection(*param.Ascending))
	case "m", "move":
		return braidMoveDispatch(bc, w, bi, param, newDefault)
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

func braidMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BraidInteractorIF, param BraidWebInput, newDefault func(string) *BraidWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	// 札は基礎札にしか動かせない。唯一の例外が捨て札→ヘルパーで、これは基礎札に
	// 出せない札を退避させるための枠。枠同士やブレイド→ヘルパーの手は存在しない。
	case fromZone == "braid" && toZone == "foundation":
		bc.writePresenterResponse(w, bi.MoveBraidToFoundation())
	case fromZone == "field" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveFieldToFoundation(*param.From.Col))
	case fromZone == "helper" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveHelperToFoundation(*param.From.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, bi.MoveWasteToFoundation())
	case fromZone == "waste" && toZone == "helper":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveWasteToHelper(*param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
