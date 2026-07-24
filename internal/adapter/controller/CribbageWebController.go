//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CribbageWebInput クリベッジWebインプット
type CribbageWebInput struct {
	BaseWebInput
	CardIndex   *int               `json:"cardIndex,omitempty"`
	CardIndices []int              `json:"cardIndices,omitempty"`
	Config      *CribbageWebConfig `json:"config,omitempty"`
}

// CribbageWebConfig クリベッジWeb設定
type CribbageWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// CribbageWebOutputPlayer クリベッジWebアウトプットプレイヤー
type CribbageWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
}

// CribbageWebOutputScoreDetail スコア詳細のアウトプット
type CribbageWebOutputScoreDetail struct {
	Fifteens int `json:"fifteens"`
	Pairs    int `json:"pairs"`
	Runs     int `json:"runs"`
	Flush    int `json:"flush"`
	Nobs     int `json:"nobs"`
	Total    int `json:"total"`
}

// CribbageWebOutput クリベッジWebアウトプット
type CribbageWebOutput struct {
	Players          []*CribbageWebOutputPlayer       `json:"players"`
	Phase            int                              `json:"phase"`
	RoundNumber      int                              `json:"roundNumber"`
	CurrentPlayerIdx int                              `json:"currentPlayerIdx"`
	DealerIdx        int                              `json:"dealerIdx"`
	Crib             []*WebOutputCard                 `json:"crib"`
	Starter          *WebOutputCard                   `json:"starter"`
	PegCount         int                              `json:"pegCount"`
	PegPlayedCards   []*WebOutputCard                 `json:"pegPlayedCards"`
	ShowPhaseStep    int                              `json:"showPhaseStep"`
	HandScoreDetails [3]*CribbageWebOutputScoreDetail `json:"handScoreDetails"`
	GameEndFlag      bool                             `json:"gameEndFlag"`
	WinnerIdx        int                              `json:"winnerIdx"`
	WebOutputBase
	Config CribbageWebOutputConfig `json:"config"`
}

// CribbageWebOutputConfig クリベッジ設定アウトプット
type CribbageWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a CribbageConfig from the nested web config, applying bounds checking.
func (c *CribbageWebConfig) ToConfig() domain.CribbageConfig {
	cfg := domain.DefaultCribbageConfig()
	cfg.CpuDifficulty = domain.CribbageCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.CribbageCpuDifficultyEasy), int(domain.CribbageCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a CribbageConfig from the web input.
func (p CribbageWebInput) ToConfig() domain.CribbageConfig {
	return configOrDefault(p.Config, (*CribbageWebConfig).ToConfig, domain.DefaultCribbageConfig())
}

// CribbageWebController クリベッジWebコントローラークラス
type CribbageWebController = GameWebController[usecase.CribbageInteractorIF, CribbageWebInput, *CribbageWebOutput]

// NewCribbageWebController and NewCribbageWebControllerWithProvider are
// the standard and provider-backed constructors for CribbageWebController.
var NewCribbageWebController, NewCribbageWebControllerWithProvider = webControllerPair[usecase.CribbageInteractorIF, CribbageWebInput, *CribbageWebOutput](
	newCribbageDefaultOutput, cribbageDispatch,
)

func newCribbageDefaultOutput(msg string) *CribbageWebOutput {
	return &CribbageWebOutput{
		Players:        make([]*CribbageWebOutputPlayer, 0),
		Crib:           make([]*WebOutputCard, 0),
		PegPlayedCards: make([]*WebOutputCard, 0),
		WinnerIdx:      -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func cribbageDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CribbageInteractorIF, param CribbageWebInput, newDefault func(string) *CribbageWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "d", "discard":
		bc.writePresenterResponse(w, ci.Discard(param.CardIndices))
	case "c", "cut":
		bc.writePresenterResponse(w, ci.Cut())
	case "p", "peg":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Peg(*param.CardIndex))
	case "go":
		bc.writePresenterResponse(w, ci.Go())
	case "sn", "shownext":
		bc.writePresenterResponse(w, ci.ShowNext())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
