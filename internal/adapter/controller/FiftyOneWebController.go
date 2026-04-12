package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FiftyOneWebInput フィフティワンWebインプット
type FiftyOneWebInput struct {
	BaseWebInput
	HandIdx  *int                    `json:"handIdx,omitempty"`
	TableIdx *int                    `json:"tableIdx,omitempty"`
	Config   *FiftyOneWebInputConfig `json:"config,omitempty"`
}

// FiftyOneWebInputConfig フィフティワン設定インプット
type FiftyOneWebInputConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// FiftyOneWebOutputPlayer フィフティワンWebアウトプットプレイヤー
type FiftyOneWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	Score     int              `json:"score"`
}

// FiftyOneWebOutput フィフティワンWebアウトプット
type FiftyOneWebOutput struct {
	Players       []*FiftyOneWebOutputPlayer `json:"players"`
	TableCards    []*WebOutputCard           `json:"tableCards"`
	Phase         int                        `json:"phase"`
	CurrentTurn   int                        `json:"currentTurn"`
	GameEndFlag   bool                       `json:"gameEndFlag"`
	WinnerIdx     int                        `json:"winnerIdx"`
	TurnNumber    int                        `json:"turnNumber"`
	StopCallerIdx int                        `json:"stopCallerIdx"`
	LastAction    string                     `json:"lastAction"`
	LastHandIdx   int                        `json:"lastHandIdx"`
	LastTableIdx  int                        `json:"lastTableIdx"`
	WebOutputBase
	Config FiftyOneWebOutputConfig `json:"config"`
}

// FiftyOneWebOutputConfig フィフティワン設定アウトプット
type FiftyOneWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// FiftyOneWebController フィフティワンWebコントローラークラス
type FiftyOneWebController = GameWebController[usecase.FiftyOneInteractorIF, FiftyOneWebInput, *FiftyOneWebOutput]

// NewFiftyOneWebController and NewFiftyOneWebControllerWithProvider are
// the standard and provider-backed constructors for FiftyOneWebController.
var NewFiftyOneWebController, NewFiftyOneWebControllerWithProvider = webControllerPair[usecase.FiftyOneInteractorIF, FiftyOneWebInput, *FiftyOneWebOutput](
	newFiftyOneDefaultOutput, fiftyOneDispatch,
)

func newFiftyOneDefaultOutput(msg string) *FiftyOneWebOutput {
	return &FiftyOneWebOutput{
		Players:       make([]*FiftyOneWebOutputPlayer, 0),
		TableCards:    make([]*WebOutputCard, 0),
		WinnerIdx:     -1,
		StopCallerIdx: -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func fiftyOneDispatch(bc *baseController, w http.ResponseWriter, fi usecase.FiftyOneInteractorIF, param FiftyOneWebInput, newDefault func(string) *FiftyOneWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		cfg := buildFiftyOneConfig(fi.GetConfig(), param.Config)
		bc.writePresenterResponse(w, fi.Reset(cfg))
	case "play":
		if param.HandIdx == nil || param.TableIdx == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: handIdx and tableIdx are required."))
			return true
		}
		bc.writePresenterResponse(w, fi.ExchangeOne(*param.HandIdx, *param.TableIdx))
	case "exchangeall":
		bc.writePresenterResponse(w, fi.ExchangeAll())
	case "stop":
		bc.writePresenterResponse(w, fi.Stop())
	default:
		return dispatchLog(param.Command, bc, w, fi.ActionLog)
	}
	return true
}

func buildFiftyOneConfig(current domain.FiftyOneConfig, input *FiftyOneWebInputConfig) domain.FiftyOneConfig {
	if input == nil {
		return current
	}
	cfg := current
	if input.CpuDifficulty != nil {
		cfg.CpuDifficulty = domain.FiftyOneCpuDifficulty(*input.CpuDifficulty)
	}
	return cfg
}
