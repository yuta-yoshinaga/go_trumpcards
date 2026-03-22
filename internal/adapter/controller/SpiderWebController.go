package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// SpiderWebInput スパイダーソリティアWebインプット
type SpiderWebInput struct {
	BaseWebInput
	From   *SpiderWebZone   `json:"from,omitempty"`
	To     *SpiderWebZone   `json:"to,omitempty"`
	Config *SpiderWebConfig `json:"config,omitempty"`
}

// SpiderWebConfig 設定
type SpiderWebConfig struct {
	Difficulty *int `json:"difficulty,omitempty"`
}

// SpiderWebZone ゾーン指定
type SpiderWebZone struct {
	Zone      string `json:"zone"`
	Col       *int   `json:"col,omitempty"`
	CardIndex *int   `json:"cardIndex,omitempty"`
}

// SpiderWebOutputTableauCard タブローカード出力
type SpiderWebOutputTableauCard struct {
	Card   *WebOutputCard `json:"card"`
	FaceUp bool           `json:"faceUp"`
}

// SpiderWebOutputHint ヒント出力
type SpiderWebOutputHint struct {
	FromCol   int `json:"fromCol"`
	CardIndex int `json:"cardIndex"`
	ToCol     int `json:"toCol"`
}

// SpiderWebOutput スパイダーソリティアWebアウトプット
type SpiderWebOutput struct {
	Tableau        [][]*SpiderWebOutputTableauCard `json:"tableau"`
	StockCount     int                             `json:"stockCount"`
	CompletedSuits int                             `json:"completedSuits"`
	Phase          int                             `json:"phase"`
	MoveCount      int                             `json:"moveCount"`
	CanUndo        bool                            `json:"canUndo"`
	IsStalemate    bool                            `json:"isStalemate"`
	Score          int                             `json:"score"`
	Difficulty     int                             `json:"difficulty"`
	Hint           *SpiderWebOutputHint            `json:"hint,omitempty"`
	WebOutputBase
}

// SpiderWebController スパイダーソリティアWebコントローラークラス
type SpiderWebController = GameWebController[usecase.SpiderInteractorIF, SpiderWebInput, *SpiderWebOutput]

// NewSpiderWebController コンストラクタ
func NewSpiderWebController(factory func() usecase.SpiderInteractorIF) *SpiderWebController {
	return NewGameWebController(factory, newSpiderDefaultOutput, spiderDispatch)
}

func newSpiderDefaultOutput(msg string) *SpiderWebOutput {
	return &SpiderWebOutput{
		Tableau:       make([][]*SpiderWebOutputTableauCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func spiderDispatch(bc *baseController, w rest.ResponseWriter, si usecase.SpiderInteractorIF, param SpiderWebInput, newDefault func(string) *SpiderWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			cfg := domain.SpiderConfig{}
			if param.Config.Difficulty != nil {
				cfg.Difficulty = domain.SpiderDifficulty(*param.Config.Difficulty)
			}
			bc.writePresenterResponse(w, si.ResetWithConfig(cfg))
		} else {
			bc.writePresenterResponse(w, si.Reset())
		}
	case "d", "deal":
		bc.writePresenterResponse(w, si.Deal())
	case "m", "move":
		return spiderMoveDispatch(bc, w, si, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	case "h", "hint":
		bc.writePresenterResponse(w, si.Hint())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, si.AutoComplete())
	case "log", "l":
		bc.writePresenterResponse(w, si.ActionLog())
	case "u", "undo":
		bc.writePresenterResponse(w, si.Undo())
	default:
		return false
	}
	return true
}

func spiderMoveDispatch(bc *baseController, w rest.ResponseWriter, si usecase.SpiderInteractorIF, param SpiderWebInput, newDefault func(string) *SpiderWebOutput) bool {
	if param.From == nil || param.To == nil {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from and to are required."))
		return true
	}
	fromZone := param.From.Zone
	toZone := param.To.Zone

	if fromZone == "tableau" && toZone == "tableau" {
		if param.From.Col == nil || param.From.CardIndex == nil || param.To.Col == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: from.col, from.cardIndex, to.col are required."))
			return true
		}
		bc.writePresenterResponse(w, si.MoveTableauToTableau(*param.From.Col, *param.From.CardIndex, *param.To.Col))
	} else {
		bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: invalid move zones. Only tableau to tableau is supported."))
	}
	return true
}
