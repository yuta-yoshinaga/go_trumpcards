//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MissMilliganWebInput ミス・ミリガン Web インプット
type MissMilliganWebInput struct {
	BaseWebInput
	From *MissMilliganWebZone `json:"from,omitempty"`
	To   *MissMilliganWebZone `json:"to,omitempty"`
}

// MissMilliganWebZone ゾーン指定。Zone は "tableau" / "waived" / "foundation"。
type MissMilliganWebZone struct {
	Zone string `json:"zone"`
	Col  *int   `json:"col,omitempty"`
	// CardIndex は移動する連番グループの先頭。省略すると最上段 1 枚だけを動かす。
	CardIndex *int `json:"cardIndex,omitempty"`
}

// MissMilliganWebOutputTableauCard タブローカード出力
type MissMilliganWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// MissMilliganWebOutputHint ヒント出力
type MissMilliganWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToIdx     int    `json:"toIdx"`
}

// MissMilliganWebOutput ミス・ミリガン Web アウトプット
type MissMilliganWebOutput struct {
	Tableau    [][]*MissMilliganWebOutputTableauCard `json:"tableau"`
	StockCount int                                   `json:"stockCount"`
	Foundation [][]*WebOutputCard                    `json:"foundation"`
	// Waived は保持中の札。空でない間は配り足しもウェイブもできない。
	Waived []*WebOutputCard `json:"waived"`
	// CanWaive は今ウェイブできるか（山札を使い切り、まだ何も保持していない）。
	CanWaive bool                       `json:"canWaive"`
	Hint     *MissMilliganWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// MissMilliganWebController ミス・ミリガン Web コントローラークラス
type MissMilliganWebController = GameWebController[usecase.MissMilliganInteractorIF, MissMilliganWebInput, *MissMilliganWebOutput]

// NewMissMilliganWebController and NewMissMilliganWebControllerWithProvider are
// the standard and provider-backed constructors for MissMilliganWebController.
var NewMissMilliganWebController, NewMissMilliganWebControllerWithProvider = webControllerPair[usecase.MissMilliganInteractorIF, MissMilliganWebInput, *MissMilliganWebOutput](
	newMissMilliganDefaultOutput, missMilliganDispatch,
)

func newMissMilliganDefaultOutput(msg string) *MissMilliganWebOutput {
	return &MissMilliganWebOutput{
		Tableau:       make([][]*MissMilliganWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		Waived:        make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func missMilliganDispatch(bc *baseController, w http.ResponseWriter, mi usecase.MissMilliganInteractorIF, param MissMilliganWebInput, newDefault func(string) *MissMilliganWebOutput) bool {
	switch param.Command {
	case "d", "deal":
		bc.writePresenterResponse(w, mi.Deal())
	case "m", "move":
		return missMilliganMoveDispatch(bc, w, mi, param, newDefault)
	case "wv", "waive":
		if !requireParam(bc, w, newDefault, param.From == nil || param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		cardIndex := -1
		if param.From.CardIndex != nil {
			cardIndex = *param.From.CardIndex
		}
		bc.writePresenterResponse(w, mi.Waive(*param.From.Col, cardIndex))
	case "g", "giveup":
		bc.writePresenterResponse(w, mi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, mi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, mi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, mi.Reset, mi.Hint, mi.ActionLog)
	}
	return true
}

func missMilliganMoveDispatch(bc *baseController, w http.ResponseWriter, mi usecase.MissMilliganInteractorIF, param MissMilliganWebInput, newDefault func(string) *MissMilliganWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "waived" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.PlaceWaived(*param.To.Col))
	case fromZone == "waived" && toZone == "foundation":
		bc.writePresenterResponse(w, mi.MoveWaivedToFoundation())
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		// cardIndex は連番グループの先頭。省略時は -1 = 最上段 1 枚。
		cardIndex := -1
		if param.From.CardIndex != nil {
			cardIndex = *param.From.CardIndex
		}
		bc.writePresenterResponse(w, mi.MoveTableauToTableau(*param.From.Col, cardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
