//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TrappolaWebInput トラッポラのWebインプット
type TrappolaWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *TrappolaWebConfig `json:"config,omitempty"`
}

// TrappolaWebConfig トラッポラのWeb設定
type TrappolaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// TrappolaWebOutputPlayer トラッポラのWebアウトプットプレイヤー
type TrappolaWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	TeamID     int              `json:"teamId"`
}

// TrappolaWebOutput トラッポラのWebアウトプット
type TrappolaWebOutput struct {
	Players          []*TrappolaWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TrickNumber      int                        `json:"trickNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	LastTrick        []*WebOutputTrickCard      `json:"lastTrick"`
	LastTrickWinner  int                        `json:"lastTrickWinner"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	TeamScores       []int                      `json:"teamScores"`
	TeamRoundThirds  []int                      `json:"teamRoundThirds"`
	PlayableIndices  []int                      `json:"playableIndices"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerTeam       int                        `json:"winnerTeam"`
	Hint             *WebOutputCardHint         `json:"hint,omitempty"`
	WebOutputBase
	Config TrappolaWebOutputConfig `json:"config"`
}

// TrappolaWebOutputConfig トラッポラの設定アウトプット
type TrappolaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a TrappolaConfig from the nested web config, applying bounds checking.
func (c *TrappolaWebConfig) ToConfig() domain.TrappolaConfig {
	cfg := domain.DefaultTrappolaConfig()
	cfg.CpuDifficulty = domain.TrappolaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.TrappolaCpuDifficultyEasy), int(domain.TrappolaCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 100000)
	return cfg
}

// ToConfig builds a TrappolaConfig from the web input.
func (p TrappolaWebInput) ToConfig() domain.TrappolaConfig {
	return configOrDefault(p.Config, (*TrappolaWebConfig).ToConfig, domain.DefaultTrappolaConfig())
}

// TrappolaWebController トラッポラのWebコントローラークラス
type TrappolaWebController = GameWebController[usecase.TrappolaInteractorIF, TrappolaWebInput, *TrappolaWebOutput]

// NewTrappolaWebController and NewTrappolaWebControllerWithProvider are
// the standard and provider-backed constructors for TrappolaWebController.
var NewTrappolaWebController, NewTrappolaWebControllerWithProvider = webControllerPair[usecase.TrappolaInteractorIF, TrappolaWebInput, *TrappolaWebOutput](
	newTrappolaDefaultOutput, trappolaDispatch,
)

func newTrappolaDefaultOutput(msg string) *TrappolaWebOutput {
	return &TrappolaWebOutput{
		Players:         make([]*TrappolaWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrick:       make([]*WebOutputTrickCard, 0),
		LastTrickWinner: -1,
		TeamScores:      make([]int, 0),
		TeamRoundThirds: make([]int, 0),
		PlayableIndices: make([]int, 0),
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func trappolaDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TrappolaInteractorIF, param TrappolaWebInput, newDefault func(string) *TrappolaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ti.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
