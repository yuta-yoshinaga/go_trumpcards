//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CucumberWebInput キューカンバーWebインプット
type CucumberWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *CucumberWebConfig `json:"config,omitempty"`
}

// CucumberWebConfig キューカンバーWeb設定
type CucumberWebConfig struct {
	PlayerCnt   *int `json:"playerCnt,omitempty"`
	TargetScore *int `json:"targetScore,omitempty"`
}

// CucumberWebOutputPlayer キューカンバーWebアウトプットプレイヤー
type CucumberWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Penalty は失点。**少ないほうが良い。**
	Penalty int `json:"penalty"`
}

// CucumberWebOutputHint ヒント出力
type CucumberWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// CucumberWebOutput キューカンバーWebアウトプット
type CucumberWebOutput struct {
	Players    []*CucumberWebOutputPlayer `json:"players"`
	Phase      int                        `json:"phase"`
	ValidPlays []int                      `json:"validPlays"`
	// HighestInTrick はいま出ている最高ランク (0 = まだ無い)。**超えるかどうかの基準。**
	HighestInTrick int                   `json:"highestInTrick"`
	CurrentTrick   []*WebOutputTrickCard `json:"currentTrick"`
	// Forced は「更新できないので低い札に決まっている」場面か。
	Forced             bool                   `json:"forced"`
	CurrentPlayerIdx   int                    `json:"currentPlayerIdx"`
	LeadPlayerIdx      int                    `json:"leadPlayerIdx"`
	TrickNumber        int                    `json:"trickNumber"`
	RoundNumber        int                    `json:"roundNumber"`
	LastTrickWinnerIdx int                    `json:"lastTrickWinnerIdx"`
	LastPenalty        int                    `json:"lastPenalty"`
	GameEndFlag        bool                   `json:"gameEndFlag"`
	WinnerIdx          int                    `json:"winnerIdx"`
	Hint               *CucumberWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config CucumberWebOutputConfig `json:"config"`
}

// CucumberWebOutputConfig キューカンバー設定アウトプット
type CucumberWebOutputConfig struct {
	PlayerCnt   int `json:"playerCnt"`
	TargetScore int `json:"targetScore"`
}

// ToConfig builds a CucumberConfig from the nested web config, applying bounds checking.
func (c *CucumberWebConfig) ToConfig() domain.CucumberConfig {
	cfg := domain.DefaultCucumberConfig()
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.CucumberPlayerCntMin, domain.CucumberPlayerCntMax, cfg.PlayerCnt)
	cfg.TargetScore = webutil.BoundedIntPtr(c.TargetScore,
		domain.CucumberTargetScoreMin, domain.CucumberTargetScoreMax, cfg.TargetScore)
	return cfg
}

// ToConfig builds a CucumberConfig from the web input.
func (p CucumberWebInput) ToConfig() domain.CucumberConfig {
	return configOrDefault(p.Config, (*CucumberWebConfig).ToConfig, domain.DefaultCucumberConfig())
}

// CucumberWebController キューカンバーWebコントローラークラス
type CucumberWebController = GameWebController[usecase.CucumberInteractorIF, CucumberWebInput, *CucumberWebOutput]

// NewCucumberWebController and NewCucumberWebControllerWithProvider are the
// standard and provider-backed constructors for CucumberWebController.
var NewCucumberWebController, NewCucumberWebControllerWithProvider = webControllerPair[usecase.CucumberInteractorIF, CucumberWebInput, *CucumberWebOutput](
	newCucumberDefaultOutput, cucumberDispatch,
)

func newCucumberDefaultOutput(msg string) *CucumberWebOutput {
	return &CucumberWebOutput{
		Players:            make([]*CucumberWebOutputPlayer, 0),
		ValidPlays:         make([]int, 0),
		CurrentTrick:       make([]*WebOutputTrickCard, 0),
		LastTrickWinnerIdx: -1,
		WinnerIdx:          -1,
		WebOutputBase:      WebOutputBase{Message: msg},
	}
}

func cucumberDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CucumberInteractorIF, param CucumberWebInput, newDefault func(string) *CucumberWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, ci.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ci.Hint, ci.ActionLog)
	}
	return true
}
