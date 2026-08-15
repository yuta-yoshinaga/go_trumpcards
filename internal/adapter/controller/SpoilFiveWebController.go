//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpoilFiveWebInput スポイル・ファイブのWebインプット
type SpoilFiveWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *SpoilFiveWebConfig `json:"config,omitempty"`
}

// SpoilFiveWebConfig スポイル・ファイブのWeb設定
type SpoilFiveWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// SpoilFiveWebOutputPlayer スポイル・ファイブのWebアウトプットプレイヤー
type SpoilFiveWebOutputPlayer struct {
	ID          int              `json:"id"`
	IsHuman     bool             `json:"isHuman"`
	CardCount   int              `json:"cardCount"`
	Cards       []*WebOutputCard `json:"cards"`
	TrickCount  int              `json:"trickCount"`
	Score       int              `json:"score"`
	RoundTricks int              `json:"roundTricks"`
}

// SpoilFiveWebOutput スポイル・ファイブのWebアウトプット
type SpoilFiveWebOutput struct {
	Players          []*SpoilFiveWebOutputPlayer `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	DealerIdx        int                         `json:"dealerIdx"`
	TrumpSuit        int                         `json:"trumpSuit"`
	Pot              int                         `json:"pot"`
	RoundWinnerIdx   int                         `json:"roundWinnerIdx"`
	CurrentTrick     []*WebOutputTrickCard       `json:"currentTrick"`
	PlayableIndices  []int                       `json:"playableIndices"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerPlayer     int                         `json:"winnerPlayer"`
	IsHumanTurn      bool                        `json:"isHumanTurn"`
	Hint             *WebOutputCardHint          `json:"hint,omitempty"`
	WebOutputBase
	Config SpoilFiveWebOutputConfig `json:"config"`
}

// SpoilFiveWebOutputConfig スポイル・ファイブの設定アウトプット
type SpoilFiveWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a SpoilFiveConfig from the nested web config, applying bounds checking.
func (c *SpoilFiveWebConfig) ToConfig() domain.SpoilFiveConfig {
	cfg := domain.DefaultSpoilFiveConfig()
	cfg.CpuDifficulty = domain.SpoilFiveCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.SpoilFiveCpuDifficultyEasy), int(domain.SpoilFiveCpuDifficultyHard), int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a SpoilFiveConfig from the web input.
func (p SpoilFiveWebInput) ToConfig() domain.SpoilFiveConfig {
	return configOrDefault(p.Config, (*SpoilFiveWebConfig).ToConfig, domain.DefaultSpoilFiveConfig())
}

// SpoilFiveWebController スポイル・ファイブのWebコントローラークラス
type SpoilFiveWebController = GameWebController[usecase.SpoilFiveInteractorIF, SpoilFiveWebInput, *SpoilFiveWebOutput]

// NewSpoilFiveWebController and NewSpoilFiveWebControllerWithProvider are
// the standard and provider-backed constructors for SpoilFiveWebController.
var NewSpoilFiveWebController, NewSpoilFiveWebControllerWithProvider = webControllerPair[usecase.SpoilFiveInteractorIF, SpoilFiveWebInput, *SpoilFiveWebOutput](
	newSpoilFiveDefaultOutput, spoilFiveDispatch,
)

func newSpoilFiveDefaultOutput(msg string) *SpoilFiveWebOutput {
	return &SpoilFiveWebOutput{
		Players:         make([]*SpoilFiveWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		RoundWinnerIdx:  -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func spoilFiveDispatch(bc *baseController, w http.ResponseWriter, di usecase.SpoilFiveInteractorIF, param SpoilFiveWebInput, newDefault func(string) *SpoilFiveWebOutput) bool {
	return dispatchTrickPlay(param.Command, bc, w, trickPlayFns{
		resetWithConfig: func() string { return di.ResetWithConfig(param.ToConfig()) },
		play:            di.Play,
		nextTrick:       di.NextTrick,
		nextRound:       di.NextRound,
		hint:            di.Hint,
		actionLog:       di.ActionLog,
	}, param.CardIndex, newDefault)
}
