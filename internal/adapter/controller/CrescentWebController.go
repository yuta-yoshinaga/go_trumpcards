package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CrescentWebInput クレセント・ソリティアの Web インプット。
type CrescentWebInput struct {
	BaseWebInput
	From *CrescentWebZone `json:"from,omitempty"`
	To   *CrescentWebZone `json:"to,omitempty"`
}

// CrescentWebZone ゾーン指定。
//
//	Zone: "tableau" または "foundation"。
//	Col:  タブローなら列番号、ファンデーションならファンデーション ID (0..7)。
type CrescentWebZone struct {
	Zone string `json:"zone"`
	Col  *int   `json:"col,omitempty"`
}

// CrescentWebOutputTableauCard タブローカード出力。
type CrescentWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// CrescentWebOutputHint ヒント出力。
type CrescentWebOutputHint struct {
	FromCol int    `json:"fromCol"`
	ToZone  string `json:"toZone"`
	ToCol   int    `json:"toCol"`
	Redeal  bool   `json:"redeal"`
}

// CrescentWebOutput クレセント・ソリティアの Web アウトプット。
type CrescentWebOutput struct {
	Tableau          [][]*CrescentWebOutputTableauCard `json:"tableau"`
	Foundation       [][]*WebOutputCard                `json:"foundation"`
	RedealsRemaining int                               `json:"redealsRemaining"`
	Hint             *CrescentWebOutputHint            `json:"hint,omitempty"`
	SolitaireWebOutputBase
	WebOutputBase
}

// CrescentWebController クレセント・ソリティアの Web コントローラー。
type CrescentWebController = GameWebController[usecase.CrescentInteractorIF, CrescentWebInput, *CrescentWebOutput]

// NewCrescentWebController / NewCrescentWebControllerWithProvider は標準とプロバイダ版のコンストラクタ。
var NewCrescentWebController, NewCrescentWebControllerWithProvider = webControllerPair[usecase.CrescentInteractorIF, CrescentWebInput, *CrescentWebOutput](
	newCrescentDefaultOutput, crescentDispatch,
)

func newCrescentDefaultOutput(msg string) *CrescentWebOutput {
	return &CrescentWebOutput{
		Tableau:       make([][]*CrescentWebOutputTableauCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func crescentDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CrescentInteractorIF, param CrescentWebInput, newDefault func(string) *CrescentWebOutput) bool {
	switch param.Command {
	case "m", "move":
		return crescentMoveDispatch(bc, w, ci, param, newDefault)
	case "r", "redeal":
		bc.writePresenterResponse(w, ci.Redeal())
	case "g", "giveup":
		bc.writePresenterResponse(w, ci.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ci.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ci.Undo())
	case "undo_n":
		if param.N == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: n is required."))
			return true
		}
		bc.writePresenterResponse(w, ci.UndoN(*param.N))
	default:
		return dispatchResetHintAndLog(param.Command, bc, w, ci.Reset, ci.Hint, ci.ActionLog)
	}
	return true
}

func crescentMoveDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CrescentInteractorIF, param CrescentWebInput, newDefault func(string) *CrescentWebOutput) bool {
	if param.From == nil || param.To == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from and to are required."))
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "tableau" && toZone == "tableau":
		if param.From.Col == nil || param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col and to.col are required."))
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToTableau(*param.From.Col, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if param.From.Col == nil || param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col and to.col are required."))
			return true
		}
		bc.writePresenterResponse(w, ci.MoveTableauToFoundation(*param.From.Col, *param.To.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
