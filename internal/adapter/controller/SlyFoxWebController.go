//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SlyFoxWebInput スライ・フォックス Web インプット
type SlyFoxWebInput struct {
	BaseWebInput
	From *SlyFoxWebZone `json:"from,omitempty"`
	To   *SlyFoxWebZone `json:"to,omitempty"`
}

// SlyFoxWebZone ゾーン指定。Zone は "tableau" / "foundation"。
// 捨て札は無く、めくった札は配る先を直接指定するので waste ゾーンは無い。
type SlyFoxWebZone struct {
	Zone string `json:"zone"`
	// Idx はリザーブ枠（0..19）または基礎札（0..7）。
	//
	// フロントエンド (`SlyFoxMoveZone`)、CLI パーサ、openapi.yaml がすべて
	// `idx` を送るので、ここも `idx` でなければならない。Congress 系は `col`
	// を使っており、雛形からコピーするとここだけ食い違う。
	Idx *int `json:"idx,omitempty"`
}

// SlyFoxWebOutputHint ヒント出力
type SlyFoxWebOutputHint struct {
	FromZone string `json:"fromZone"`
	FromIdx  int    `json:"fromIdx"`
	ToZone   string `json:"toZone"`
	ToIdx    int    `json:"toIdx"`
}

// SlyFoxWebOutput スライ・フォックス Web アウトプット
type SlyFoxWebOutput struct {
	Tableau    [][]*WebOutputCard `json:"tableau"`
	Foundation [][]*WebOutputCard `json:"foundation"`
	// FoundationAscending は基礎札ごとの積む向き。true が A→K、false が K→A。
	// 添字から推測させると、並びを変えたときに表示だけが静かにずれる。
	FoundationAscending []bool `json:"foundationAscending"`
	StockCount          int    `json:"stockCount"`
	// DealtThisCycle この周でリザーブに置いた枚数。**UI はこれを見て、まだ
	// 開いていないリザーブを無効化する。**見ないと必ず拒まれる手を出せてしまう。
	DealtThisCycle int `json:"dealtThisCycle"`
	// DealCycle 1 周の枚数（定数だが、UI が「あと何枚」を出すのに要る）。
	DealCycle int `json:"dealCycle"`
	// ReserveLocked この周を配り切るまでリザーブが閉じているか。
	ReserveLocked bool                 `json:"reserveLocked"`
	Hint          *SlyFoxWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// SlyFoxWebController スライ・フォックス Web コントローラークラス
type SlyFoxWebController = GameWebController[usecase.SlyFoxInteractorIF, SlyFoxWebInput, *SlyFoxWebOutput]

// NewSlyFoxWebController and NewSlyFoxWebControllerWithProvider are the
// standard and provider-backed constructors for SlyFoxWebController.
var NewSlyFoxWebController, NewSlyFoxWebControllerWithProvider = webControllerPair[usecase.SlyFoxInteractorIF, SlyFoxWebInput, *SlyFoxWebOutput](
	newSlyFoxDefaultOutput, slyFoxDispatch,
)

func newSlyFoxDefaultOutput(msg string) *SlyFoxWebOutput {
	return &SlyFoxWebOutput{
		Tableau:             make([][]*WebOutputCard, 0),
		Foundation:          make([][]*WebOutputCard, 0),
		FoundationAscending: make([]bool, 0),
		DealCycle:           domain.SlyFoxDealCycle,
		WebOutputBase:       WebOutputBase{Message: msg},
	}
}

func slyFoxDispatch(bc *baseController, w http.ResponseWriter, ci usecase.SlyFoxInteractorIF, param SlyFoxWebInput, newDefault func(string) *SlyFoxWebOutput) bool {
	switch param.Command {
	case "d", "deal":
		// **配り先は必須。**捨て札が無いので、行き先を決めずには配れない。
		if !requireParam(bc, w, newDefault, param.To == nil, "param error: to is required.") {
			return true
		}
		if param.To.Zone == "foundation" {
			if !requireParam(bc, w, newDefault, param.To.Idx == nil, "param error: to.idx is required.") {
				return true
			}
			bc.writePresenterResponse(w, ci.DealToFoundation(*param.To.Idx))
			return true
		}
		if !requireParam(bc, w, newDefault, param.To.Idx == nil, "param error: to.idx is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.DealToPile(*param.To.Idx))
	case "m", "move":
		return slyFoxMoveDispatch(bc, w, ci, param, newDefault)
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

func slyFoxMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.SlyFoxInteractorIF, param SlyFoxWebInput, newDefault func(string) *SlyFoxWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Idx == nil, "param error: from.idx is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToFoundation(*param.From.Idx))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
