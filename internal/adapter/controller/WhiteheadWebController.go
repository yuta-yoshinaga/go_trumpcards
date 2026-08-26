//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WhiteheadWebInput ホワイトヘッドWebインプット
type WhiteheadWebInput struct {
	BaseWebInput
	From   *WhiteheadWebZone   `json:"from,omitempty"`
	To     *WhiteheadWebZone   `json:"to,omitempty"`
	Config *WhiteheadWebConfig `json:"config,omitempty"`
}

// WhiteheadWebConfig 設定
type WhiteheadWebConfig struct {
	DrawCount   *int `json:"drawCount,omitempty"`
	ScoringMode *int `json:"scoringMode,omitempty"`
}

// WhiteheadWebZone ゾーン指定
type WhiteheadWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// WhiteheadWebOutputTableauCard タブローカード出力
type WhiteheadWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// WhiteheadWebOutputHint ヒント出力
type WhiteheadWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// WhiteheadWebOutput ホワイトヘッドWebアウトプット
type WhiteheadWebOutput struct {
	Tableau     [][]*WhiteheadWebOutputTableauCard `json:"tableau"`
	StockCount  int                                `json:"stockCount"`
	Waste       []*WebOutputCard                   `json:"waste"`
	Foundation  [][]*WebOutputCard                 `json:"foundation"`
	DrawCount   int                                `json:"drawCount"`
	Score       int                                `json:"score"`
	ScoringMode int                                `json:"scoringMode"`
	Hint        *WhiteheadWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// WhiteheadWebController ホワイトヘッドWebコントローラークラス
type WhiteheadWebController = GameWebController[usecase.WhiteheadInteractorIF, WhiteheadWebInput, *WhiteheadWebOutput]

// NewWhiteheadWebController and NewWhiteheadWebControllerWithProvider are
// the standard and provider-backed constructors for WhiteheadWebController.
var NewWhiteheadWebController, NewWhiteheadWebControllerWithProvider = webControllerPair[usecase.WhiteheadInteractorIF, WhiteheadWebInput, *WhiteheadWebOutput](
	newWhiteheadDefaultOutput, whiteheadDispatch,
)

func newWhiteheadDefaultOutput(msg string) *WhiteheadWebOutput {
	return &WhiteheadWebOutput{
		Tableau:       make([][]*WhiteheadWebOutputTableauCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func whiteheadDispatch(bc *baseController, w http.ResponseWriter, ki usecase.WhiteheadInteractorIF, param WhiteheadWebInput, newDefault func(string) *WhiteheadWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			cfg := domain.WhiteheadConfig{}
			if param.Config.DrawCount != nil {
				cfg.DrawCount = *param.Config.DrawCount
			}
			if param.Config.ScoringMode != nil {
				cfg.ScoringMode = domain.WhiteheadScoringMode(*param.Config.ScoringMode)
			}
			bc.writePresenterResponse(w, ki.ResetWithConfig(cfg))
		} else {
			bc.writePresenterResponse(w, ki.Reset())
		}
	case "d", "draw":
		bc.writePresenterResponse(w, ki.Draw())
	case "m", "move":
		return whiteheadMoveDispatch(bc, w, ki, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ki.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ki.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ki.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, ki.UndoN(*param.N))
	default:
		return dispatchHintAndLog(param.Command, bc, w, ki.Hint, ki.ActionLog)
	}
	return true
}

func whiteheadMoveDispatch(bc *baseController, w http.ResponseWriter, ki usecase.WhiteheadInteractorIF, param WhiteheadWebInput, newDefault func(string) *WhiteheadWebOutput) bool {
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
		bc.writePresenterResponse(w, ki.MoveWasteToTableau(*param.To.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, ki.MoveWasteToFoundation())
	case fromZone == "tableau" && toZone == "tableau":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil, "param error: from.col, from.cardIndex, to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, ki.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil, "param error: from.col is required.") {
			return true
		}
		bc.writePresenterResponse(w, ki.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
