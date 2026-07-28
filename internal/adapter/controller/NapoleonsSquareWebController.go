//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NapoleonsSquareWebInput ナポレオンズ・スクエア Web インプット
type NapoleonsSquareWebInput struct {
	BaseWebInput
	From *NapoleonsSquareWebZone `json:"from,omitempty"`
	To   *NapoleonsSquareWebZone `json:"to,omitempty"`
}

// NapoleonsSquareWebZone ゾーン指定。Zone は "waste" / "tableau" / "foundation"。
type NapoleonsSquareWebZone struct {
	Zone string `json:"zone"`
	Col  *int   `json:"col,omitempty"`
	// CardIndex は移動する連番グループの先頭。省略すると最上段 1 枚だけを動かす。
	CardIndex *int `json:"cardIndex,omitempty"`
}

// NapoleonsSquareWebOutputTableauCard タブローカード出力
type NapoleonsSquareWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// NapoleonsSquareWebOutputHint ヒント出力
type NapoleonsSquareWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// NapoleonsSquareWebOutput ナポレオンズ・スクエア Web アウトプット
type NapoleonsSquareWebOutput struct {
	Tableau    [][]*NapoleonsSquareWebOutputTableauCard `json:"tableau"`
	StockCount int                                      `json:"stockCount"`
	Waste      []*WebOutputCard                         `json:"waste"`
	Foundation [][]*WebOutputCard                       `json:"foundation"`
	Hint       *NapoleonsSquareWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// NapoleonsSquareWebController ナポレオンズ・スクエア Web コントローラークラス
type NapoleonsSquareWebController = GameWebController[usecase.NapoleonsSquareInteractorIF, NapoleonsSquareWebInput, *NapoleonsSquareWebOutput]

// NewNapoleonsSquareWebController and NewNapoleonsSquareWebControllerWithProvider are
// the standard and provider-backed constructors for NapoleonsSquareWebController.
var NewNapoleonsSquareWebController, NewNapoleonsSquareWebControllerWithProvider = webControllerPair[usecase.NapoleonsSquareInteractorIF, NapoleonsSquareWebInput, *NapoleonsSquareWebOutput](
	newNapoleonsSquareDefaultOutput, napoleonsSquareDispatch,
)

func newNapoleonsSquareDefaultOutput(msg string) *NapoleonsSquareWebOutput {
	return &NapoleonsSquareWebOutput{
		Tableau:       make([][]*NapoleonsSquareWebOutputTableauCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func napoleonsSquareDispatch(bc *baseController, w http.ResponseWriter, ni usecase.NapoleonsSquareInteractorIF, param NapoleonsSquareWebInput, newDefault func(string) *NapoleonsSquareWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ni.Draw())
	case "m", "move":
		return napoleonsSquareMoveDispatch(bc, w, ni, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ni.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ni.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ni.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ni.Reset, ni.Hint, ni.ActionLog)
	}
	return true
}

func napoleonsSquareMoveDispatch(bc *baseController, w http.ResponseWriter, ni usecase.NapoleonsSquareInteractorIF, param NapoleonsSquareWebInput, newDefault func(string) *NapoleonsSquareWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "waste" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.MoveWasteToTableau(*param.To.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, ni.MoveWasteToFoundation())
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		// cardIndex は連番グループの先頭。省略時は最上段 1 枚だけを動かす、という
		// 意味に倒す（インデックスは presenter が返すので、クライアントが自前で
		// 計算する必要はない）。
		cardIndex := -1
		if param.From.CardIndex != nil {
			cardIndex = *param.From.CardIndex
		}
		bc.writePresenterResponse(w, ni.MoveTableauToTableau(*param.From.Col, cardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
