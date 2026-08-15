//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SedmaWebInput セドマのWebインプット
type SedmaWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *SedmaWebConfig `json:"config,omitempty"`
}

// SedmaWebConfig セドマのWeb設定
type SedmaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// SedmaWebOutputPlayer セドマのWebアウトプットプレイヤー
type SedmaWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	TeamScore  int              `json:"teamScore"`
}

// SedmaWebOutput セドマのWebアウトプット
type SedmaWebOutput struct {
	Players          []*SedmaWebOutputPlayer  `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	TeamScores       [domain.SedmaTeamCnt]int `json:"teamScores"`
	RoundCardPoints  [domain.SedmaTeamCnt]int `json:"roundCardPoints"`
	PlayableIndices  []int                    `json:"playableIndices"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerTeam       int                      `json:"winnerTeam"`
	IsHumanTurn      bool                     `json:"isHumanTurn"`
	Hint             *WebOutputCardHint       `json:"hint,omitempty"`
	WebOutputBase
	Config SedmaWebOutputConfig `json:"config"`
}

// SedmaWebOutputConfig セドマの設定アウトプット
type SedmaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a SedmaConfig from the nested web config, applying bounds checking.
func (c *SedmaWebConfig) ToConfig() domain.SedmaConfig {
	cfg := domain.DefaultSedmaConfig()
	cfg.CpuDifficulty = domain.SedmaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.SedmaCpuDifficultyEasy), int(domain.SedmaCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a SedmaConfig from the web input.
func (p SedmaWebInput) ToConfig() domain.SedmaConfig {
	return configOrDefault(p.Config, (*SedmaWebConfig).ToConfig, domain.DefaultSedmaConfig())
}

// SedmaWebController セドマのWebコントローラークラス
type SedmaWebController = GameWebController[usecase.SedmaInteractorIF, SedmaWebInput, *SedmaWebOutput]

// NewSedmaWebController and NewSedmaWebControllerWithProvider are
// the standard and provider-backed constructors for SedmaWebController.
var NewSedmaWebController, NewSedmaWebControllerWithProvider = webControllerPair[usecase.SedmaInteractorIF, SedmaWebInput, *SedmaWebOutput](
	newSedmaDefaultOutput, sedmaDispatch,
)

func newSedmaDefaultOutput(msg string) *SedmaWebOutput {
	return &SedmaWebOutput{
		Players:         make([]*SedmaWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func sedmaDispatch(bc *baseController, w http.ResponseWriter, di usecase.SedmaInteractorIF, param SedmaWebInput, newDefault func(string) *SedmaWebOutput) bool {
	return dispatchTrickPlay(param.Command, bc, w, trickPlayFns{
		resetWithConfig: func() string { return di.ResetWithConfig(param.ToConfig()) },
		play:            di.Play,
		nextTrick:       di.NextTrick,
		nextRound:       di.NextRound,
		hint:            di.Hint,
		actionLog:       di.ActionLog,
	}, param.CardIndex, newDefault)
}
