//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ManilleWebInput マニーユのWebインプット
type ManilleWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *ManilleWebConfig `json:"config,omitempty"`
}

// ManilleWebConfig マニーユのWeb設定
type ManilleWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// ManilleWebOutputPlayer マニーユのWebアウトプットプレイヤー
type ManilleWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	TeamScore  int              `json:"teamScore"`
}

// ManilleWebOutput マニーユのWebアウトプット
type ManilleWebOutput struct {
	Players          []*ManilleWebOutputPlayer  `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TrickNumber      int                        `json:"trickNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	DealerIdx        int                        `json:"dealerIdx"`
	TrumpSuit        int                        `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	TeamScores       [domain.ManilleTeamCnt]int `json:"teamScores"`
	RoundCardPoints  [domain.ManilleTeamCnt]int `json:"roundCardPoints"`
	PlayableIndices  []int                      `json:"playableIndices"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerTeam       int                        `json:"winnerTeam"`
	IsHumanTurn      bool                       `json:"isHumanTurn"`
	Hint             *WebOutputCardHint         `json:"hint,omitempty"`
	WebOutputBase
	Config ManilleWebOutputConfig `json:"config"`
}

// ManilleWebOutputConfig マニーユの設定アウトプット
type ManilleWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a ManilleConfig from the nested web config, applying bounds checking.
func (c *ManilleWebConfig) ToConfig() domain.ManilleConfig {
	cfg := domain.DefaultManilleConfig()
	cfg.CpuDifficulty = domain.ManilleCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.ManilleCpuDifficultyEasy), int(domain.ManilleCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a ManilleConfig from the web input.
func (p ManilleWebInput) ToConfig() domain.ManilleConfig {
	return configOrDefault(p.Config, (*ManilleWebConfig).ToConfig, domain.DefaultManilleConfig())
}

// ManilleWebController マニーユのWebコントローラークラス
type ManilleWebController = GameWebController[usecase.ManilleInteractorIF, ManilleWebInput, *ManilleWebOutput]

// NewManilleWebController and NewManilleWebControllerWithProvider are
// the standard and provider-backed constructors for ManilleWebController.
var NewManilleWebController, NewManilleWebControllerWithProvider = webControllerPair[usecase.ManilleInteractorIF, ManilleWebInput, *ManilleWebOutput](
	newManilleDefaultOutput, manilleDispatch,
)

func newManilleDefaultOutput(msg string) *ManilleWebOutput {
	return &ManilleWebOutput{
		Players:         make([]*ManilleWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func manilleDispatch(bc *baseController, w http.ResponseWriter, di usecase.ManilleInteractorIF, param ManilleWebInput, newDefault func(string) *ManilleWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
