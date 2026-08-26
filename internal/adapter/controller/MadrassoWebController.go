//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MadrassoWebInput マドラッソのWebインプット
type MadrassoWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *MadrassoWebConfig `json:"config,omitempty"`
}

// MadrassoWebConfig マドラッソのWeb設定
type MadrassoWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// MadrassoWebOutputPlayer マドラッソのWebアウトプットプレイヤー
type MadrassoWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	TeamID     int              `json:"teamId"`
}

// MadrassoWebOutput マドラッソのWebアウトプット
type MadrassoWebOutput struct {
	Players          []*MadrassoWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TrickNumber      int                        `json:"trickNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	LastTrick        []*WebOutputTrickCard      `json:"lastTrick"`
	LastTrickWinner  int                        `json:"lastTrickWinner"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	TeamScores       []int                      `json:"teamScores"`
	TeamRoundPoints  []int                      `json:"teamRoundPoints"`
	// TrumpSuit は配りで決まった切り札スート (-1=未確定)。
	TrumpSuit       int                `json:"trumpSuit"`
	PlayableIndices []int              `json:"playableIndices"`
	GameEndFlag     bool               `json:"gameEndFlag"`
	WinnerTeam      int                `json:"winnerTeam"`
	Hint            *WebOutputCardHint `json:"hint,omitempty"`
	WebOutputBase
	Config MadrassoWebOutputConfig `json:"config"`
}

// MadrassoWebOutputConfig マドラッソの設定アウトプット
type MadrassoWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a MadrassoConfig from the nested web config, applying bounds checking.
func (c *MadrassoWebConfig) ToConfig() domain.MadrassoConfig {
	cfg := domain.DefaultMadrassoConfig()
	cfg.CpuDifficulty = domain.MadrassoCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.MadrassoCpuDifficultyEasy), int(domain.MadrassoCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 100000)
	return cfg
}

// ToConfig builds a MadrassoConfig from the web input.
func (p MadrassoWebInput) ToConfig() domain.MadrassoConfig {
	return configOrDefault(p.Config, (*MadrassoWebConfig).ToConfig, domain.DefaultMadrassoConfig())
}

// MadrassoWebController マドラッソのWebコントローラークラス
type MadrassoWebController = GameWebController[usecase.MadrassoInteractorIF, MadrassoWebInput, *MadrassoWebOutput]

// NewMadrassoWebController and NewMadrassoWebControllerWithProvider are
// the standard and provider-backed constructors for MadrassoWebController.
var NewMadrassoWebController, NewMadrassoWebControllerWithProvider = webControllerPair[usecase.MadrassoInteractorIF, MadrassoWebInput, *MadrassoWebOutput](
	newMadrassoDefaultOutput, madrassoDispatch,
)

func newMadrassoDefaultOutput(msg string) *MadrassoWebOutput {
	return &MadrassoWebOutput{
		Players:         make([]*MadrassoWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrick:       make([]*WebOutputTrickCard, 0),
		LastTrickWinner: -1,
		TeamScores:      make([]int, 0),
		TeamRoundPoints: make([]int, 0),
		PlayableIndices: make([]int, 0),
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func madrassoDispatch(bc *baseController, w http.ResponseWriter, ti usecase.MadrassoInteractorIF, param MadrassoWebInput, newDefault func(string) *MadrassoWebOutput) bool {
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
