package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// KlondikeWebInput クロンダイクWebインプット
type KlondikeWebInput struct {
	BaseWebInput
	From *KlondikeWebZone `json:"from,omitempty"`
	To   *KlondikeWebZone `json:"to,omitempty"`
}

// KlondikeWebZone ゾーン指定
type KlondikeWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// KlondikeWebOutputTableauCard タブローカード出力
type KlondikeWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// KlondikeWebOutputHint ヒント出力
type KlondikeWebOutputHint struct {
	FromZone  string `json:"fromZone"`
	FromCol   int    `json:"fromCol"`
	CardIndex int    `json:"cardIndex"`
	ToZone    string `json:"toZone"`
	ToCol     int    `json:"toCol"`
}

// KlondikeWebOutput クロンダイクWebアウトプット
type KlondikeWebOutput struct {
	Tableau       [][]*KlondikeWebOutputTableauCard `json:"tableau"`
	StockCount    int                               `json:"stockCount"`
	Waste         []*WebOutputCard                  `json:"waste"`
	Foundation    [][]*WebOutputCard                `json:"foundation"`
	Phase         int                               `json:"phase"`
	MoveCount     int                               `json:"moveCount"`
	Message       string                            `json:"message"`
	MessageCode   string                            `json:"messageCode,omitempty"`
	MessageParams map[string]string                 `json:"messageParams,omitempty"`
	Hint          *KlondikeWebOutputHint            `json:"hint,omitempty"`
}

// KlondikeWebController クロンダイクWebコントローラークラス
type KlondikeWebController = GameWebController[usecase.KlondikeInteractorIF, KlondikeWebInput, *KlondikeWebOutput]

// NewKlondikeWebController コンストラクタ
func NewKlondikeWebController(factory func() usecase.KlondikeInteractorIF) *KlondikeWebController {
	return NewGameWebController(factory, newKlondikeDefaultOutput, klondikeDispatch)
}

func newKlondikeDefaultOutput(msg string) *KlondikeWebOutput {
	return &KlondikeWebOutput{
		Tableau:    make([][]*KlondikeWebOutputTableauCard, 0),
		Waste:      make([]*WebOutputCard, 0),
		Foundation: make([][]*WebOutputCard, 0),
		Message:    msg,
	}
}

func klondikeDispatch(bc *baseController, w rest.ResponseWriter, ki usecase.KlondikeInteractorIF, param KlondikeWebInput, newDefault func(string) *KlondikeWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ki.Reset())
	case "d", "draw":
		bc.writePresenterResponse(w, ki.Draw())
	case "m", "move":
		return klondikeMoveDispatch(bc, w, ki, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ki.GiveUp())
	case "h", "hint":
		bc.writePresenterResponse(w, ki.Hint())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ki.AutoComplete())
	case "log", "l":
		bc.writePresenterResponse(w, ki.ActionLog())
	default:
		return false
	}
	return true
}

func klondikeMoveDispatch(bc *baseController, w rest.ResponseWriter, ki usecase.KlondikeInteractorIF, param KlondikeWebInput, newDefault func(string) *KlondikeWebOutput) bool {
	if param.From == nil || param.To == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from and to are required."))
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	switch {
	case fromZone == "waste" && toZone == "tableau":
		if param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: to.col is required."))
			return true
		}
		bc.writePresenterResponse(w, ki.MoveWasteToTableau(*param.To.Col))
	case fromZone == "waste" && toZone == "foundation":
		bc.writePresenterResponse(w, ki.MoveWasteToFoundation())
	case fromZone == "tableau" && toZone == "tableau":
		if param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col, from.cardIndex, to.col are required."))
			return true
		}
		bc.writePresenterResponse(w, ki.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	case fromZone == "tableau" && toZone == "foundation":
		if param.From.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col is required."))
			return true
		}
		bc.writePresenterResponse(w, ki.MoveTableauToFoundation(*param.From.Col))
	default:
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones."))
	}
	return true
}
