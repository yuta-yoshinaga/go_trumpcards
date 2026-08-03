//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AmericanToadWebInput アメリカン・トード Web インプット
type AmericanToadWebInput struct {
	BaseWebInput
	From *AmericanToadWebZone `json:"from,omitempty"`
	To   *AmericanToadWebZone `json:"to,omitempty"`
}

// AmericanToadWebZone ゾーン指定。Zone は "reserve" / "waste" / "tableau" / "foundation"。
type AmericanToadWebZone struct {
	Zone string `json:"zone"`
	// Col はタブロー列（0..7）。リザーブ・捨て札・基礎札では不要。
	Col *int `json:"col,omitempty"`
	// CardIndex は移動する連番グループの先頭。省略すると最上段 1 枚だけを動かす。
	CardIndex *int `json:"cardIndex,omitempty"`
}

// AmericanToadWebOutputTableauCard タブローカード出力
type AmericanToadWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// AmericanToadWebOutputHint ヒント出力
type AmericanToadWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromIdx   int    `json:"fromIdx"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToIdx     int    `json:"toIdx"`
}

// AmericanToadWebOutput アメリカン・トード Web アウトプット
type AmericanToadWebOutput struct {
	Reserve    []*WebOutputCard                      `json:"reserve"`
	Tableau    [][]*AmericanToadWebOutputTableauCard `json:"tableau"`
	Foundation [][]*WebOutputCard                    `json:"foundation"`
	StockCount int                                   `json:"stockCount"`
	Waste      []*WebOutputCard                      `json:"waste"`
	// BaseRank は 8 つの基礎札が始まるランク。
	BaseRank int `json:"baseRank"`
	// PassesUsed は山札を通した回数。CanRedeal と併せて残りのめくり直しが分かる。
	PassesUsed int `json:"passesUsed"`
	// CanRedeal が真なら、山札が空でも捨て札を戻してもう一巡できる。
	CanRedeal bool                       `json:"canRedeal"`
	Hint      *AmericanToadWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// AmericanToadWebController アメリカン・トード Web コントローラークラス
type AmericanToadWebController = GameWebController[usecase.AmericanToadInteractorIF, AmericanToadWebInput, *AmericanToadWebOutput]

// NewAmericanToadWebController and NewAmericanToadWebControllerWithProvider are
// the standard and provider-backed constructors for AmericanToadWebController.
var NewAmericanToadWebController, NewAmericanToadWebControllerWithProvider = webControllerPair[usecase.AmericanToadInteractorIF, AmericanToadWebInput, *AmericanToadWebOutput](
	newAmericanToadDefaultOutput, americanToadDispatch,
)

func newAmericanToadDefaultOutput(msg string) *AmericanToadWebOutput {
	return &AmericanToadWebOutput{
		Reserve:       make([]*WebOutputCard, 0),
		Tableau:       make([][]*AmericanToadWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func americanToadDispatch(bc *baseController, w http.ResponseWriter, ai usecase.AmericanToadInteractorIF, param AmericanToadWebInput, newDefault func(string) *AmericanToadWebOutput) bool {
	switch param.Command {
	case "d", "draw":
		bc.writePresenterResponse(w, ai.Draw())
	case "m", "move":
		return americanToadMoveDispatch(bc, w, ai, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ai.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ai.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ai.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ai.Reset, ai.Hint, ai.ActionLog)
	}
	return true
}

func americanToadMoveDispatch(bc *baseController, w http.ResponseWriter, ai usecase.AmericanToadInteractorIF, param AmericanToadWebInput, newDefault func(string) *AmericanToadWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "reserve" && toZone == "foundation":
		bc.writePresenterResponse(w, ai.MoveReserveToFoundation())
	case fromZone == "reserve" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.MoveReserveToTableau(*param.To.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, ai.MoveWasteToFoundation())
	case fromZone == "waste" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.To.Col == nil, "param error: to.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.MoveWasteToTableau(*param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		// cardIndex は連番グループの先頭。省略時は -1 = 最上段 1 枚。
		cardIndex := -1
		if param.From.CardIndex != nil {
			cardIndex = *param.From.CardIndex
		}
		bc.writePresenterResponse(w, ai.MoveTableauToTableau(*param.From.Col, cardIndex, *param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
