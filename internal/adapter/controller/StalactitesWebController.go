//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// StalactitesWebInput フリーセルWebインプット
type StalactitesWebInput struct {
	BaseWebInput
	From *StalactitesWebZone `json:"from,omitempty"`
	To   *StalactitesWebZone `json:"to,omitempty"`
}

// StalactitesWebZone ゾーン指定
type StalactitesWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	Cell      *int   `json:"cell,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// StalactitesWebOutputHint ヒント出力
type StalactitesWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// StalactitesWebOutput フリーセルWebアウトプット
type StalactitesWebOutput struct {
	Tableau    [][]*WebOutputCard `json:"tableau"`
	Cells      []*WebOutputCard   `json:"cells"`
	Foundation [][]*WebOutputCard `json:"foundation"`
	// MaxMovableCards / MaxMovableCardsToEmptyColumn はドメインが決めた上限を
	// そのまま運ぶ。フロントで数え直すと、空き列を経由地に使えない分の差
	// (ドメインの maxMovableCards(toCol)) が抜け、動かせない束を「動かせる」と
	// 表示してしまう (#5975)。
	// BaseRank はファンデーションの開始ランク。Stalactites は Ace 固定では
	// なく配りごとに変わるので、これを送らないとフロントは「その札が組札に
	// 置けるか」を判定できない。
	BaseRank                     int                       `json:"baseRank"`
	MaxMovableCards              int                       `json:"maxMovableCards"`
	MaxMovableCardsToEmptyColumn int                       `json:"maxMovableCardsToEmptyColumn"`
	Hint                         *StalactitesWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// StalactitesWebController フリーセルWebコントローラークラス
type StalactitesWebController = GameWebController[usecase.StalactitesInteractorIF, StalactitesWebInput, *StalactitesWebOutput]

// NewStalactitesWebController and NewStalactitesWebControllerWithProvider are
// the standard and provider-backed constructors for StalactitesWebController.
var NewStalactitesWebController, NewStalactitesWebControllerWithProvider = webControllerPair[usecase.StalactitesInteractorIF, StalactitesWebInput, *StalactitesWebOutput](
	newStalactitesDefaultOutput, stalactitesDispatch,
)

func newStalactitesDefaultOutput(msg string) *StalactitesWebOutput {
	return &StalactitesWebOutput{
		Tableau:       make([][]*WebOutputCard, 0),
		Cells:         make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func stalactitesDispatch(bc *baseController, w http.ResponseWriter, fi usecase.StalactitesInteractorIF, param StalactitesWebInput, newDefault func(string) *StalactitesWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return stalactitesMoveDispatch(bc, w, fi, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, fi.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, fi.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, fi.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, fi.Reset, fi.Hint, fi.ActionLog)
	}
	return true
}

func stalactitesMoveDispatch(bc *baseController, w http.ResponseWriter, fi usecase.StalactitesInteractorIF, param StalactitesWebInput, newDefault func(string) *StalactitesWebOutput) bool {
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
		bc.writePresenterResponse(w, fi.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.MoveTableauToFoundation(*param.From.Col))
	case fromZone == "tableau" && toZone == "stalactites":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Cell == nil, "param error: from.col and to.cell are required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.MoveTableauToStalactites(*param.From.Col, *param.To.Cell))
	case fromZone == "stalactites" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Cell == nil || param.To.Col == nil, "param error: from.cell and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.MoveStalactitesToTableau(*param.From.Cell, *param.To.Col))
	case fromZone == "stalactites" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Cell == nil, "param error: from.cell is required.") {
			return true
		}
		bc.writePresenterResponse(w, fi.MoveStalactitesToFoundation(*param.From.Cell))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
