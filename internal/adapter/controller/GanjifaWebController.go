//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GanjifaWebInput ガンジファのWebインプット
type GanjifaWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *GanjifaWebConfig `json:"config,omitempty"`
}

// GanjifaWebConfig ガンジファのWeb設定
type GanjifaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// GanjifaWebOutputPlayer ガンジファのWebアウトプットプレイヤー
type GanjifaWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
}

// GanjifaWebOutput ガンジファのWebアウトプット
type GanjifaWebOutput struct {
	Players          []*GanjifaWebOutputPlayer    `json:"players"`
	Phase            int                          `json:"phase"`
	RoundNumber      int                          `json:"roundNumber"`
	TrickNumber      int                          `json:"trickNumber"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                          `json:"leadPlayerIdx"`
	DealerIdx        int                          `json:"dealerIdx"`
	TrumpSuit        int                          `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard        `json:"currentTrick"`
	PlayerScores     [domain.GanjifaPlayerCnt]int `json:"playerScores"`
	RoundTricks      [domain.GanjifaPlayerCnt]int `json:"roundTricks"`
	PlayableIndices  []int                        `json:"playableIndices"`
	GameEndFlag      bool                         `json:"gameEndFlag"`
	WinnerPlayer     int                          `json:"winnerPlayer"`
	IsHumanTurn      bool                         `json:"isHumanTurn"`
	Hint             *WebOutputCardHint           `json:"hint,omitempty"`
	WebOutputBase
	Config GanjifaWebOutputConfig `json:"config"`
}

// GanjifaWebOutputConfig ガンジファの設定アウトプット
type GanjifaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds a GanjifaConfig from the nested web config, applying bounds checking.
func (c *GanjifaWebConfig) ToConfig() domain.GanjifaConfig {
	cfg := domain.DefaultGanjifaConfig()
	cfg.CpuDifficulty = domain.GanjifaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.GanjifaCpuDifficultyEasy), int(domain.GanjifaCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 1000000)
	return cfg
}

// ToConfig builds a GanjifaConfig from the web input.
func (p GanjifaWebInput) ToConfig() domain.GanjifaConfig {
	return configOrDefault(p.Config, (*GanjifaWebConfig).ToConfig, domain.DefaultGanjifaConfig())
}

// GanjifaWebController ガンジファのWebコントローラークラス
type GanjifaWebController = GameWebController[usecase.GanjifaInteractorIF, GanjifaWebInput, *GanjifaWebOutput]

// NewGanjifaWebController and NewGanjifaWebControllerWithProvider are
// the standard and provider-backed constructors for GanjifaWebController.
var NewGanjifaWebController, NewGanjifaWebControllerWithProvider = webControllerPair[usecase.GanjifaInteractorIF, GanjifaWebInput, *GanjifaWebOutput](
	newGanjifaDefaultOutput, ganjifaDispatch,
)

func newGanjifaDefaultOutput(msg string) *GanjifaWebOutput {
	return &GanjifaWebOutput{
		Players:         make([]*GanjifaWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func ganjifaDispatch(bc *baseController, w http.ResponseWriter, di usecase.GanjifaInteractorIF, param GanjifaWebInput, newDefault func(string) *GanjifaWebOutput) bool {
	return dispatchTrickPlay(param.Command, bc, w, trickPlayFns{
		resetWithConfig: func() string { return di.ResetWithConfig(param.ToConfig()) },
		play:            di.Play,
		nextTrick:       di.NextTrick,
		nextRound:       di.NextRound,
		hint:            di.Hint,
		actionLog:       di.ActionLog,
	}, param.CardIndex, newDefault)
}
