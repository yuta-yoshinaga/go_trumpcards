//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MrsMopWebInput ミセス・モップソリティアWebインプット
type MrsMopWebInput struct {
	BaseWebInput
	From   *MrsMopWebZone   `json:"from,omitempty"`
	To     *MrsMopWebZone   `json:"to,omitempty"`
	Config *MrsMopWebConfig `json:"config,omitempty"`
}

// MrsMopWebConfig 設定
type MrsMopWebConfig struct {
	Difficulty *int `json:"difficulty,omitempty"`
}

// MrsMopWebZone ゾーン指定
type MrsMopWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// MrsMopWebOutputTableauCard タブローカード出力
type MrsMopWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// MrsMopWebOutputHint ヒント出力
type MrsMopWebOutputHint struct {
	FromCol   int `json:"fromCol"`
	CardIndex int `json:"cardIndex"`
	ToCol     int `json:"toCol"`
}

// MrsMopWebOutput ミセス・モップソリティアWebアウトプット
type MrsMopWebOutput struct {
	Tableau        [][]*MrsMopWebOutputTableauCard `json:"tableau"`
	StockCount     int                             `json:"stockCount"`
	CompletedSuits int                             `json:"completedSuits"`
	Score          int                             `json:"score"`
	Difficulty     int                             `json:"difficulty"`
	Hint           *MrsMopWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// MrsMopWebController ミセス・モップソリティアWebコントローラークラス
type MrsMopWebController = GameWebController[usecase.MrsMopInteractorIF, MrsMopWebInput, *MrsMopWebOutput]

// NewMrsMopWebController and NewMrsMopWebControllerWithProvider are
// the standard and provider-backed constructors for MrsMopWebController.
var NewMrsMopWebController, NewMrsMopWebControllerWithProvider = webControllerPair[usecase.MrsMopInteractorIF, MrsMopWebInput, *MrsMopWebOutput](
	newMrsMopDefaultOutput, mrsMopDispatch,
)

func newMrsMopDefaultOutput(msg string) *MrsMopWebOutput {
	return &MrsMopWebOutput{
		Tableau:       make([][]*MrsMopWebOutputTableauCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func mrsMopDispatch(bc *baseController, w http.ResponseWriter, si usecase.MrsMopInteractorIF, param MrsMopWebInput, newDefault func(string) *MrsMopWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			cfg := domain.MrsMopConfig{}
			if param.Config.Difficulty != nil {
				cfg.Difficulty = domain.MrsMopDifficulty(*param.Config.Difficulty)
			}
			bc.writePresenterResponse(w, si.ResetWithConfig(cfg))
		} else {
			bc.writePresenterResponse(w, si.Reset())
		}
	case "m", "move":
		return mrsMopMoveDispatch(bc, w, si, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, si.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, si.Undo())
	case "undo_n":
		if !requireParam(bc, w, newDefault, param.N == nil, "param error: n is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.UndoN(*param.N))
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}

func mrsMopMoveDispatch(bc *baseController, w http.ResponseWriter, si usecase.MrsMopInteractorIF, param MrsMopWebInput, newDefault func(string) *MrsMopWebOutput) bool {
	if !requireParam(bc, w, newDefault, param.From == nil || param.To == nil, "param error: from and to are required.") {
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	if fromZone == "tableau" && toZone == "tableau" {
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil, "param error: from.col, from.cardIndex, to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, si.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	} else {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones. Only tableau to tableau is supported."))
	}
	return true
}
