//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// StHelenaWebInput セント・ヘレナ・ソリティアの Web インプット。
type StHelenaWebInput struct {
	BaseWebInput
	From *StHelenaWebZone `json:"from,omitempty"`
	To   *StHelenaWebZone `json:"to,omitempty"`
}

// StHelenaWebZone ゾーン指定。
//
//	Zone: "tableau" または "foundation"。
//	Col:  タブローなら列番号、ファンデーションならファンデーション ID (0..7)。
type StHelenaWebZone struct {
	Zone string `json:"zone"`
	Col  *int   `json:"col,omitempty"`
}

// StHelenaWebOutputTableauCard タブローカード出力。
type StHelenaWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// StHelenaWebOutputHint ヒント出力。
type StHelenaWebOutputHint struct {
	FromCol int    `json:"fromCol"`
	ToZone  string `json:"toZone"`
	ToCol   int    `json:"toCol"`
	Redeal  bool   `json:"redeal"`
}

// StHelenaWebOutput セント・ヘレナ・ソリティアの Web アウトプット。
type StHelenaWebOutput struct {
	Tableau          [][]*StHelenaWebOutputTableauCard `json:"tableau"`
	Foundation       [][]*WebOutputCard                `json:"foundation"`
	RedealsRemaining int                               `json:"redealsRemaining"`
	// RestrictionsActive 初回の配りの送り先制限が効いているか。**UI はこれを見て
	// 送れない組札を無効化する。**見ないと、押した瞬間にサーバが必ず拒む。
	RestrictionsActive bool                   `json:"restrictionsActive"`
	Hint               *StHelenaWebOutputHint `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// StHelenaWebController セント・ヘレナ・ソリティアの Web コントローラー。
type StHelenaWebController = GameWebController[usecase.StHelenaInteractorIF, StHelenaWebInput, *StHelenaWebOutput]

// NewStHelenaWebController / NewStHelenaWebControllerWithProvider は標準とプロバイダ版のコンストラクタ。
var NewStHelenaWebController, NewStHelenaWebControllerWithProvider = webControllerPair[usecase.StHelenaInteractorIF, StHelenaWebInput, *StHelenaWebOutput](
	newStHelenaDefaultOutput, stHelenaDispatch,
)

func newStHelenaDefaultOutput(msg string) *StHelenaWebOutput {
	return &StHelenaWebOutput{
		Tableau:       make([][]*StHelenaWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func stHelenaDispatch(bc *baseController, w http.ResponseWriter, ci usecase.StHelenaInteractorIF, param StHelenaWebInput, newDefault func(string) *StHelenaWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return stHelenaMoveDispatch(bc, w, ci, param, newDefault)
	case "r", "redeal":
		bc.writePresenterResponse(w, ci.Redeal())
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

func stHelenaMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.StHelenaInteractorIF, param StHelenaWebInput, newDefault func(string) *StHelenaWebOutput) bool {
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
		bc.writePresenterResponse(w, ci.MoveTableauToTableau(*param.From.Col, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if !requireParam(bc, w, newDefault, param.From.Col == nil || param.To.Col == nil, "param error: from.col and to.col are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToFoundation(*param.From.Col, *param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
