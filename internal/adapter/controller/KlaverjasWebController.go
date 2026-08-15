//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KlaverjasWebInput クラヴァヤスのWebインプット
type KlaverjasWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *KlaverjasWebConfig `json:"config,omitempty"`
}

// KlaverjasWebConfig クラヴァヤスのWeb設定
type KlaverjasWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// KlaverjasWebOutputPlayer クラヴァヤスのWebアウトプットプレイヤー
type KlaverjasWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	TeamScore  int              `json:"teamScore"`
}

// KlaverjasWebOutput クラヴァヤスのWebアウトプット
type KlaverjasWebOutput struct {
	Players          []*KlaverjasWebOutputPlayer  `json:"players"`
	Phase            int                          `json:"phase"`
	RoundNumber      int                          `json:"roundNumber"`
	TrickNumber      int                          `json:"trickNumber"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                          `json:"leadPlayerIdx"`
	DealerIdx        int                          `json:"dealerIdx"`
	TrumpSuit        int                          `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard        `json:"currentTrick"`
	TeamScores       [domain.KlaverjasTeamCnt]int `json:"teamScores"`
	RoundCardPoints  [domain.KlaverjasTeamCnt]int `json:"roundCardPoints"`
	RoundRoem        [domain.KlaverjasTeamCnt]int `json:"roundRoem"`
	PlayableIndices  []int                        `json:"playableIndices"`
	GameEndFlag      bool                         `json:"gameEndFlag"`
	WinnerTeam       int                          `json:"winnerTeam"`
	IsHumanTurn      bool                         `json:"isHumanTurn"`
	Hint             *WebOutputCardHint           `json:"hint,omitempty"`
	WebOutputBase
	Config KlaverjasWebOutputConfig `json:"config"`
}

// KlaverjasWebOutputConfig クラヴァヤスの設定アウトプット
type KlaverjasWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a KlaverjasConfig from the nested web config, applying bounds checking.
func (c *KlaverjasWebConfig) ToConfig() domain.KlaverjasConfig {
	cfg := domain.DefaultKlaverjasConfig()
	cfg.CpuDifficulty = domain.KlaverjasCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.KlaverjasCpuDifficultyEasy), int(domain.KlaverjasCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a KlaverjasConfig from the web input.
func (p KlaverjasWebInput) ToConfig() domain.KlaverjasConfig {
	return configOrDefault(p.Config, (*KlaverjasWebConfig).ToConfig, domain.DefaultKlaverjasConfig())
}

// KlaverjasWebController クラヴァヤスのWebコントローラークラス
type KlaverjasWebController = GameWebController[usecase.KlaverjasInteractorIF, KlaverjasWebInput, *KlaverjasWebOutput]

// NewKlaverjasWebController and NewKlaverjasWebControllerWithProvider are
// the standard and provider-backed constructors for KlaverjasWebController.
var NewKlaverjasWebController, NewKlaverjasWebControllerWithProvider = webControllerPair[usecase.KlaverjasInteractorIF, KlaverjasWebInput, *KlaverjasWebOutput](
	newKlaverjasDefaultOutput, klaverjasDispatch,
)

func newKlaverjasDefaultOutput(msg string) *KlaverjasWebOutput {
	return &KlaverjasWebOutput{
		Players:         make([]*KlaverjasWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func klaverjasDispatch(bc *baseController, w http.ResponseWriter, di usecase.KlaverjasInteractorIF, param KlaverjasWebInput, newDefault func(string) *KlaverjasWebOutput) bool {
	return dispatchTrickPlay(param.Command, bc, w, trickPlayFns{
		resetWithConfig: func() string { return di.ResetWithConfig(param.ToConfig()) },
		play:            di.Play,
		nextTrick:       di.NextTrick,
		nextRound:       di.NextRound,
		hint:            di.Hint,
		actionLog:       di.ActionLog,
	}, param.CardIndex, newDefault)
}
