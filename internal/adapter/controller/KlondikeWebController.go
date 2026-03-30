package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KlondikeWebInput クロンダイクWebインプット
type KlondikeWebInput struct {
	BaseWebInput
	From   *KlondikeWebZone   `json:"from,omitempty"`
	To     *KlondikeWebZone   `json:"to,omitempty"`
	Config *KlondikeWebConfig `json:"config,omitempty"`
}

// KlondikeWebConfig 設定
type KlondikeWebConfig struct {
	DrawCount   *int `json:"drawCount,omitempty"`
	ScoringMode *int `json:"scoringMode,omitempty"`
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
	Tableau     [][]*KlondikeWebOutputTableauCard `json:"tableau"`
	StockCount  int                               `json:"stockCount"`
	Waste       []*WebOutputCard                  `json:"waste"`
	Foundation  [][]*WebOutputCard                `json:"foundation"`
	Phase       int                               `json:"phase"`
	MoveCount   int                               `json:"moveCount"`
	DrawCount   int                               `json:"drawCount"`
	CanUndo     bool                              `json:"canUndo"`
	IsStalemate bool                              `json:"isStalemate"`
	Score       int                               `json:"score"`
	ScoringMode int                               `json:"scoringMode"`
	Hint        *KlondikeWebOutputHint            `json:"hint,omitempty"`
	WebOutputBase
}

// KlondikeWebController クロンダイクWebコントローラークラス
type KlondikeWebController = GameWebController[usecase.KlondikeInteractorIF, KlondikeWebInput, *KlondikeWebOutput]

// NewKlondikeWebController コンストラクタ
func NewKlondikeWebController(factory func() usecase.KlondikeInteractorIF) *KlondikeWebController {
	return NewGameWebController(factory, newKlondikeDefaultOutput, klondikeDispatch)
}

// NewKlondikeWebControllerWithProvider creates a KlondikeWebController with an
// explicit SessionProvider (e.g. KV-backed for Workers).
func NewKlondikeWebControllerWithProvider(
	provider SessionProvider[usecase.KlondikeInteractorIF],
	factory func() usecase.KlondikeInteractorIF,
) *KlondikeWebController {
	return NewGameWebControllerWithProvider(provider, factory, newKlondikeDefaultOutput, klondikeDispatch)
}

func newKlondikeDefaultOutput(msg string) *KlondikeWebOutput {
	return &KlondikeWebOutput{
		Tableau:       make([][]*KlondikeWebOutputTableauCard, 0),
		Waste:         make([]*WebOutputCard, 0),
		Foundation:    make([][]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func klondikeDispatch(bc *baseController, w http.ResponseWriter, ki usecase.KlondikeInteractorIF, param KlondikeWebInput, newDefault func(string) *KlondikeWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			cfg := domain.KlondikeConfig{}
			if param.Config.DrawCount != nil {
				cfg.DrawCount = *param.Config.DrawCount
			}
			if param.Config.ScoringMode != nil {
				cfg.ScoringMode = domain.KlondikeScoringMode(*param.Config.ScoringMode)
			}
			bc.writePresenterResponse(w, ki.ResetWithConfig(cfg))
		} else {
			bc.writePresenterResponse(w, ki.Reset())
		}
	case "d", "draw":
		bc.writePresenterResponse(w, ki.Draw())
	case "m", "move":
		return klondikeMoveDispatch(bc, w, ki, param, newDefault)
	case "g", "giveup":
		bc.writePresenterResponse(w, ki.GiveUp())
	case "ac", "autocomplete":
		bc.writePresenterResponse(w, ki.AutoComplete())
	case "u", "undo":
		bc.writePresenterResponse(w, ki.Undo())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ki.Hint, ki.ActionLog)
	}
	return true
}

func klondikeMoveDispatch(bc *baseController, w http.ResponseWriter, ki usecase.KlondikeInteractorIF, param KlondikeWebInput, newDefault func(string) *KlondikeWebOutput) bool {
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
