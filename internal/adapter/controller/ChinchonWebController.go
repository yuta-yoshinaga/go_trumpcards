//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ChinchonWebInput チンチョンWebインプット
type ChinchonWebInput struct {
	BaseWebInput
	CardIndex   *int               `json:"cardIndex,omitempty"`
	CardIndices []int              `json:"cardIndices,omitempty"`
	Config      *ChinchonWebConfig `json:"config,omitempty"`
}

// ChinchonWebConfig チンチョンWeb設定
type ChinchonWebConfig struct {
	CpuDifficulty    *int `json:"cpuDifficulty,omitempty"`
	PlayerCount      *int `json:"playerCount,omitempty"`
	KnockThreshold   *int `json:"knockThreshold,omitempty"`
	EliminationLimit *int `json:"eliminationLimit,omitempty"`
}

// ChinchonWebOutputMeld メルドのアウトプット
type ChinchonWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// ChinchonWebOutputPlayer チンチョンWebアウトプットプレイヤー
type ChinchonWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	Eliminated      bool             `json:"eliminated"`
}

// ChinchonWebOutput チンチョンWebアウトプット
type ChinchonWebOutput struct {
	Players          []*ChinchonWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard             `json:"discardTop"`
	DrawPileCount    int                        `json:"drawPileCount"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	KnockerIdx       int                        `json:"knockerIdx"`
	KnockerMelds     []*ChinchonWebOutputMeld   `json:"knockerMelds"`
	WebOutputBase
	Config ChinchonWebOutputConfig `json:"config"`
}

// ChinchonWebOutputConfig チンチョン設定アウトプット
type ChinchonWebOutputConfig struct {
	CpuDifficulty    int `json:"cpuDifficulty"`
	PlayerCount      int `json:"playerCount"`
	KnockThreshold   int `json:"knockThreshold"`
	EliminationLimit int `json:"eliminationLimit"`
}

// ToConfig builds a ChinchonConfig from the nested web config, applying bounds checking.
func (c *ChinchonWebConfig) ToConfig() domain.ChinchonConfig {
	cfg := domain.DefaultChinchonConfig()
	cfg.CpuDifficulty = domain.ChinchonCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.ChinchonCpuDifficultyEasy), int(domain.ChinchonCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, 2, 4)
	webutil.ApplyBoundedInt(&cfg.KnockThreshold, c.KnockThreshold, 0, 60)
	webutil.ApplyBoundedInt(&cfg.EliminationLimit, c.EliminationLimit, 1, 1000)
	return cfg
}

// ToConfig builds a ChinchonConfig from the web input.
func (p ChinchonWebInput) ToConfig() domain.ChinchonConfig {
	return configOrDefault(p.Config, (*ChinchonWebConfig).ToConfig, domain.DefaultChinchonConfig())
}

// ChinchonWebController チンチョンWebコントローラークラス
type ChinchonWebController = GameWebController[usecase.ChinchonInteractorIF, ChinchonWebInput, *ChinchonWebOutput]

// NewChinchonWebController and NewChinchonWebControllerWithProvider are
// the standard and provider-backed constructors for ChinchonWebController.
var NewChinchonWebController, NewChinchonWebControllerWithProvider = webControllerPair[usecase.ChinchonInteractorIF, ChinchonWebInput, *ChinchonWebOutput](
	newChinchonDefaultOutput, chinchonDispatch,
)

func newChinchonDefaultOutput(msg string) *ChinchonWebOutput {
	return &ChinchonWebOutput{
		Players:       make([]*ChinchonWebOutputPlayer, 0),
		WinnerIdx:     -1,
		KnockerIdx:    -1,
		KnockerMelds:  make([]*ChinchonWebOutputMeld, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func chinchonDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ChinchonInteractorIF, param ChinchonWebInput, newDefault func(string) *ChinchonWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, ci.DrawFromDiscard())
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Discard(*param.CardIndex))
	case "k", "knock":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Knock(*param.CardIndex))
	case "lo", "layoff":
		bc.writePresenterResponse(w, ci.Layoff(param.CardIndices))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
