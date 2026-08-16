//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ThirtyOneWebInput ThirtyOne Webインプット
type ThirtyOneWebInput struct {
	BaseWebInput
	CardIndex *int                `json:"cardIndex,omitempty"`
	Config    *ThirtyOneWebConfig `json:"config,omitempty"`
}

// ThirtyOneWebConfig ThirtyOne Web設定
type ThirtyOneWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	InitialLives  *int `json:"initialLives,omitempty"`
}

// ThirtyOneWebOutputPlayer ThirtyOne Webアウトプットプレイヤー
type ThirtyOneWebOutputPlayer struct {
	ID           int              `json:"id"`
	IsHuman      bool             `json:"isHuman"`
	CardCount    int              `json:"cardCount"`
	Cards        []*WebOutputCard `json:"cards"`
	Lives        int              `json:"lives"`
	Score        int              `json:"score"`
	IsEliminated bool             `json:"isEliminated"`
}

// ThirtyOneWebOutput ThirtyOne Webアウトプット
type ThirtyOneWebOutput struct {
	Players          []*ThirtyOneWebOutputPlayer `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard              `json:"discardTop"`
	DrawPileCount    int                         `json:"drawPileCount"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerIdx        int                         `json:"winnerIdx"`
	KnockerIdx       int                         `json:"knockerIdx"`
	ThirtyOneIdx     int                         `json:"thirtyOneIdx"`
	RoundWinnerIdx   int                         `json:"roundWinnerIdx"`
	RoundLosers      []int                       `json:"roundLosers"`
	WebOutputBase
	Config ThirtyOneWebOutputConfig `json:"config"`
}

// ThirtyOneWebOutputConfig ThirtyOne設定アウトプット
type ThirtyOneWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	InitialLives  int `json:"initialLives"`
}

// ToConfig builds a ThirtyOneConfig from the nested web config, applying bounds checking.
func (c *ThirtyOneWebConfig) ToConfig() domain.ThirtyOneConfig {
	cfg := domain.DefaultThirtyOneConfig()
	cfg.CpuDifficulty = domain.ThirtyOneCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.ThirtyOneCpuDifficultyEasy), int(domain.ThirtyOneCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.InitialLives, c.InitialLives, domain.ThirtyOneMinLives, domain.ThirtyOneMaxLives)
	return cfg
}

// ToConfig builds a ThirtyOneConfig from the web input.
func (p ThirtyOneWebInput) ToConfig() domain.ThirtyOneConfig {
	return configOrDefault(p.Config, (*ThirtyOneWebConfig).ToConfig, domain.DefaultThirtyOneConfig())
}

// ThirtyOneWebController ThirtyOne Webコントローラークラス
type ThirtyOneWebController = GameWebController[usecase.ThirtyOneInteractorIF, ThirtyOneWebInput, *ThirtyOneWebOutput]

// NewThirtyOneWebController and NewThirtyOneWebControllerWithProvider are
// the standard and provider-backed constructors for ThirtyOneWebController.
var NewThirtyOneWebController, NewThirtyOneWebControllerWithProvider = webControllerPair[usecase.ThirtyOneInteractorIF, ThirtyOneWebInput, *ThirtyOneWebOutput](
	newThirtyOneDefaultOutput, thirtyOneDispatch,
)

func newThirtyOneDefaultOutput(msg string) *ThirtyOneWebOutput {
	return &ThirtyOneWebOutput{
		Players:        make([]*ThirtyOneWebOutputPlayer, 0),
		WinnerIdx:      -1,
		KnockerIdx:     -1,
		ThirtyOneIdx:   -1,
		RoundWinnerIdx: -1,
		RoundLosers:    make([]int, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func thirtyOneDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ThirtyOneInteractorIF, param ThirtyOneWebInput, newDefault func(string) *ThirtyOneWebOutput) bool {
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
		bc.writePresenterResponse(w, ci.Knock())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ci.Hint, ci.ActionLog)
	}
	return true
}
