//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KingAlbertWebInput King Albert Web インプット
type KingAlbertWebInput struct {
	BaseWebInput
	From *KingAlbertWebZone `json:"from,omitempty"`
	To   *KingAlbertWebZone `json:"to,omitempty"`
}

// KingAlbertWebZone ゾーン指定
type KingAlbertWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// KingAlbertWebOutputTableauCard タブローカード出力
type KingAlbertWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// KingAlbertWebOutputHint ヒント出力
type KingAlbertWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// KingAlbertWebOutput King Albert Web アウトプット
type KingAlbertWebOutput struct {
	Tableau    [][]*KingAlbertWebOutputTableauCard `json:"tableau"`
	Reserve    []*WebOutputCard                    `json:"reserve"`
	Foundation [][]*WebOutputCard                  `json:"foundation"`
	Hint       *KingAlbertWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// KingAlbertWebController King Albert Web コントローラークラス
type KingAlbertWebController = GameWebController[usecase.KingAlbertInteractorIF, KingAlbertWebInput, *KingAlbertWebOutput]

// NewKingAlbertWebController and NewKingAlbertWebControllerWithProvider are
// the standard and provider-backed constructors for KingAlbertWebController.
var NewKingAlbertWebController, NewKingAlbertWebControllerWithProvider = webControllerPair[usecase.KingAlbertInteractorIF, KingAlbertWebInput, *KingAlbertWebOutput](
	newKingAlbertDefaultOutput, kingAlbertDispatch,
)

func newKingAlbertDefaultOutput(msg string) *KingAlbertWebOutput {
	return &KingAlbertWebOutput{
		Tableau:       make([][]*KingAlbertWebOutputTableauCard, 0),
		Reserve:       make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func kingAlbertDispatch(bc *baseController, w http.ResponseWriter, bi usecase.KingAlbertInteractorIF, param KingAlbertWebInput, newDefault func(string) *KingAlbertWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return kingAlbertMoveDispatch(bc, w, bi, param, newDefault)
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

func kingAlbertMoveDispatch(bc *baseController, w http.ResponseWriter, bi usecase.KingAlbertInteractorIF, param KingAlbertWebInput, newDefault func(string) *KingAlbertWebOutput) bool {
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
		// King Albert only ever moves the bottom card; pass -1 so the domain
		// resolves the index from its own state.
		bc.writePresenterResponse(w, bi.MoveTableauToTableau(*param.From.Col, -1, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "reserve" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveReserveToTableau(*param.From.Col, *param.To.Col))
	case fromZone == "reserve" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.MoveReserveToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
