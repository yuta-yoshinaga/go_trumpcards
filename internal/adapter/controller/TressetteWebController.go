//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TressetteWebInput トレセッテのWebインプット
type TressetteWebInput struct {
	BaseWebInput
	CardIndex *int                `json:"cardIndex,omitempty"`
	Config    *TressetteWebConfig `json:"config,omitempty"`
}

// TressetteWebConfig トレセッテのWeb設定
type TressetteWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// TressetteWebOutputPlayer トレセッテのWebアウトプットプレイヤー
type TressetteWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	TeamID     int              `json:"teamId"`
}

// TressetteWebOutput トレセッテのWebアウトプット
type TressetteWebOutput struct {
	Players          []*TressetteWebOutputPlayer `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard       `json:"currentTrick"`
	LastTrick        []*WebOutputTrickCard       `json:"lastTrick"`
	LastTrickWinner  int                         `json:"lastTrickWinner"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	TeamScores       []int                       `json:"teamScores"`
	TeamRoundThirds  []int                       `json:"teamRoundThirds"`
	PlayableIndices  []int                       `json:"playableIndices"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerTeam       int                         `json:"winnerTeam"`
	Hint             *WebOutputCardHint          `json:"hint,omitempty"`
	WebOutputBase
	Config TressetteWebOutputConfig `json:"config"`
}

// TressetteWebOutputConfig トレセッテの設定アウトプット
type TressetteWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a TressetteConfig from the nested web config, applying bounds checking.
func (c *TressetteWebConfig) ToConfig() domain.TressetteConfig {
	cfg := domain.DefaultTressetteConfig()
	cfg.CpuDifficulty = domain.TressetteCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.TressetteCpuDifficultyEasy), int(domain.TressetteCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 100000)
	return cfg
}

// ToConfig builds a TressetteConfig from the web input.
func (p TressetteWebInput) ToConfig() domain.TressetteConfig {
	return configOrDefault(p.Config, (*TressetteWebConfig).ToConfig, domain.DefaultTressetteConfig())
}

// TressetteWebController トレセッテのWebコントローラークラス
type TressetteWebController = GameWebController[usecase.TressetteInteractorIF, TressetteWebInput, *TressetteWebOutput]

// NewTressetteWebController and NewTressetteWebControllerWithProvider are
// the standard and provider-backed constructors for TressetteWebController.
var NewTressetteWebController, NewTressetteWebControllerWithProvider = webControllerPair[usecase.TressetteInteractorIF, TressetteWebInput, *TressetteWebOutput](
	newTressetteDefaultOutput, tressetteDispatch,
)

func newTressetteDefaultOutput(msg string) *TressetteWebOutput {
	return &TressetteWebOutput{
		Players:         make([]*TressetteWebOutputPlayer, 0),
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

func tressetteDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TressetteInteractorIF, param TressetteWebInput, newDefault func(string) *TressetteWebOutput) bool {
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
