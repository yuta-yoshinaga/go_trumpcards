//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ScartoWebInput スカルト (Scarto) のWebインプット
type ScartoWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices スカルトで捨てるカードのインデックス (3 枚)
	CardIndices []int `json:"cardIndices,omitempty"`
	// Config ゲーム設定
	Config *ScartoWebConfig `json:"config,omitempty"`
}

// ScartoWebConfig スカルトのWeb設定
type ScartoWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// ScartoWebOutputPlayer スカルトのWebアウトプットプレイヤー
type ScartoWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	CardPoints int              `json:"cardPoints"`
	Score      int              `json:"score"`
	IsDealer   bool             `json:"isDealer"`
}

// ScartoWebOutput スカルトのWebアウトプット
type ScartoWebOutput struct {
	Players          []*ScartoWebOutputPlayer    `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	DealerIdx        int                         `json:"dealerIdx"`
	ScartoCount      int                         `json:"scartoCount"`
	CurrentTrick     []*WebOutputTrickCard       `json:"currentTrick"`
	PlayerScores     [domain.ScartoPlayerCnt]int `json:"playerScores"`
	DealScores       [domain.ScartoPlayerCnt]int `json:"dealScores"`
	LastTrickWinner  int                         `json:"lastTrickWinner"`
	Outcome          int                         `json:"outcome"`
	Result           int                         `json:"result"`
	PlayableIndices  []int                       `json:"playableIndices"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerPlayer     int                         `json:"winnerPlayer"`
	IsHumanTurn      bool                        `json:"isHumanTurn"`
	IsHumanScarto    bool                        `json:"isHumanScarto"`
	Hint             *WebOutputCardHint          `json:"hint,omitempty"`
	WebOutputBase
	Config ScartoWebOutputConfig `json:"config"`
}

// ScartoWebOutputConfig スカルトの設定アウトプット
type ScartoWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// ToConfig builds a ScartoConfig from the nested web config, applying bounds checking.
func (c *ScartoWebConfig) ToConfig() domain.ScartoConfig {
	cfg := domain.DefaultScartoConfig()
	cfg.CpuDifficulty = domain.ScartoCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.ScartoCpuDifficultyEasy), int(domain.ScartoCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetDeals, c.TargetDeals, 1, 1000)
	return cfg
}

// ToConfig builds a ScartoConfig from the web input.
func (p ScartoWebInput) ToConfig() domain.ScartoConfig {
	return configOrDefault(p.Config, (*ScartoWebConfig).ToConfig, domain.DefaultScartoConfig())
}

// ScartoWebController スカルトのWebコントローラークラス
type ScartoWebController = GameWebController[usecase.ScartoInteractorIF, ScartoWebInput, *ScartoWebOutput]

// NewScartoWebController and NewScartoWebControllerWithProvider are the standard
// and provider-backed constructors for ScartoWebController.
var NewScartoWebController, NewScartoWebControllerWithProvider = webControllerPair[usecase.ScartoInteractorIF, ScartoWebInput, *ScartoWebOutput](
	newScartoDefaultOutput, scartoDispatch,
)

func newScartoDefaultOutput(msg string) *ScartoWebOutput {
	return &ScartoWebOutput{
		Players:         make([]*ScartoWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func scartoDispatch(bc *baseController, w http.ResponseWriter, di usecase.ScartoInteractorIF, param ScartoWebInput, newDefault func(string) *ScartoWebOutput) bool {
	return dispatchTarotDiscardPlay(param.Command, bc, w, tarotDiscardPlayFns{
		resetWithConfig: func() string { return di.ResetWithConfig(param.ToConfig()) },
		discard:         di.Discard,
		play:            di.Play,
		nextTrick:       di.NextTrick,
		nextRound:       di.NextRound,
		hint:            di.Hint,
		actionLog:       di.ActionLog,
	}, param.CardIndices, param.CardIndex, newDefault)
}
