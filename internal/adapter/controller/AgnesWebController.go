//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AgnesWebInput アグネス・ソレルWebインプット
type AgnesWebInput struct {
	BaseWebInput
	From *AgnesWebZone `json:"from,omitempty"`
	To   *AgnesWebZone `json:"to,omitempty"`
}

// AgnesWebZone ゾーン指定
type AgnesWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// AgnesWebOutputTableauCard タブローカード出力
type AgnesWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// AgnesWebOutputHint ヒント出力
type AgnesWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// AgnesWebOutput アグネス・ソレルWebアウトプット
type AgnesWebOutput struct {
	Tableau    [][]*AgnesWebOutputTableauCard `json:"tableau"`
	StockCount int                            `json:"stockCount"`
	Foundation [][]*WebOutputCard             `json:"foundation"`
	BaseRank   int                            `json:"baseRank"`
	Phase      int                            `json:"phase"`
	MoveCount  int                            `json:"moveCount"`
	CanUndo    bool                           `json:"canUndo"`
	// IsStalemate は合法手が 1 つも無いかどうか。判定はドメインの
	// Agnes.IsStalemate() だけが持つ。以前はここに載せておらず、
	// フロントが agnesHasLegalMove() で同じ規則を実装し直していた (#5601)。
	IsStalemate bool                `json:"isStalemate"`
	Hint        *AgnesWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
}

// AgnesWebController アグネス・ソレルWebコントローラークラス
type AgnesWebController = GameWebController[usecase.AgnesInteractorIF, AgnesWebInput, *AgnesWebOutput]

// NewAgnesWebController and NewAgnesWebControllerWithProvider are
// the standard and provider-backed constructors for AgnesWebController.
var NewAgnesWebController, NewAgnesWebControllerWithProvider = webControllerPair[usecase.AgnesInteractorIF, AgnesWebInput, *AgnesWebOutput](
	newAgnesDefaultOutput, agnesDispatch,
)

func newAgnesDefaultOutput(msg string) *AgnesWebOutput {
	return &AgnesWebOutput{
		Tableau:       make([][]*AgnesWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func agnesDispatch(bc *baseController, w http.ResponseWriter, ci usecase.AgnesInteractorIF, param AgnesWebInput, newDefault func(string) *AgnesWebOutput) bool {
	switch param.Command {
	case "d", "deal":
		bc.writePresenterResponse(w, ci.DealStock())
	case "m", "move":
		return agnesMoveDispatch(bc, w, ci, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ci.GiveUp())
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

func agnesMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.AgnesInteractorIF, param AgnesWebInput, newDefault func(string) *AgnesWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil, "param error: from.col, from.cardIndex, to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
