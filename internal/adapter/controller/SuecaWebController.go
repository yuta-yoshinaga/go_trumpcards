//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SuecaWebInput スエカのWebインプット
type SuecaWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *SuecaWebConfig `json:"config,omitempty"`
}

// SuecaWebConfig スエカのWeb設定
type SuecaWebConfig struct {
	CpuDifficulty    *int `json:"cpuDifficulty,omitempty"`
	TargetGamePoints *int `json:"targetGamePoints,omitempty"`
}

// SuecaWebOutputPlayer スエカのWebアウトプットプレイヤー
type SuecaWebOutputPlayer struct {
	ID             int              `json:"id"`
	IsHuman        bool             `json:"isHuman"`
	CardCount      int              `json:"cardCount"`
	Cards          []*WebOutputCard `json:"cards"`
	TrickCount     int              `json:"trickCount"`
	TeamGamePoints int              `json:"teamGamePoints"`
}

// SuecaWebOutput スエカのWebアウトプット
type SuecaWebOutput struct {
	Players          []*SuecaWebOutputPlayer  `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	TrumpSuit        int                      `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	TeamGamePoints   [domain.SuecaTeamCnt]int `json:"teamGamePoints"`
	RoundCardPoints  [domain.SuecaTeamCnt]int `json:"roundCardPoints"`
	RoundWinnerTeam  int                      `json:"roundWinnerTeam"`
	RoundGamePoints  int                      `json:"roundGamePoints"`
	PlayableIndices  []int                    `json:"playableIndices"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerTeam       int                      `json:"winnerTeam"`
	IsHumanTurn      bool                     `json:"isHumanTurn"`
	Hint             *WebOutputCardHint       `json:"hint,omitempty"`
	WebOutputBase
	Config SuecaWebOutputConfig `json:"config"`
}

// SuecaWebOutputConfig スエカの設定アウトプット
type SuecaWebOutputConfig struct {
	CpuDifficulty    int `json:"cpuDifficulty"`
	TargetGamePoints int `json:"targetGamePoints"`
}

// ToConfig builds a SuecaConfig from the nested web config, applying bounds checking.
func (c *SuecaWebConfig) ToConfig() domain.SuecaConfig {
	cfg := domain.DefaultSuecaConfig()
	cfg.CpuDifficulty = domain.SuecaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.SuecaCpuDifficultyEasy), int(domain.SuecaCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetGamePoints, c.TargetGamePoints, 1, 1000000)
	return cfg
}

// ToConfig builds a SuecaConfig from the web input.
func (p SuecaWebInput) ToConfig() domain.SuecaConfig {
	return configOrDefault(p.Config, (*SuecaWebConfig).ToConfig, domain.DefaultSuecaConfig())
}

// SuecaWebController スエカのWebコントローラークラス
type SuecaWebController = GameWebController[usecase.SuecaInteractorIF, SuecaWebInput, *SuecaWebOutput]

// NewSuecaWebController and NewSuecaWebControllerWithProvider are
// the standard and provider-backed constructors for SuecaWebController.
var NewSuecaWebController, NewSuecaWebControllerWithProvider = webControllerPair[usecase.SuecaInteractorIF, SuecaWebInput, *SuecaWebOutput](
	newSuecaDefaultOutput, suecaDispatch,
)

func newSuecaDefaultOutput(msg string) *SuecaWebOutput {
	return &SuecaWebOutput{
		Players:         make([]*SuecaWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		WinnerTeam:      -1,
		RoundWinnerTeam: -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func suecaDispatch(bc *baseController, w http.ResponseWriter, di usecase.SuecaInteractorIF, param SuecaWebInput, newDefault func(string) *SuecaWebOutput) bool {
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
