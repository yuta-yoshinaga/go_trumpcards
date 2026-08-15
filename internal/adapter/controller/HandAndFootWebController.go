//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HandAndFootWebInput ハンドアンドフットWebインプット
type HandAndFootWebInput struct {
	BaseWebInput
	CardIndex          *int                  `json:"cardIndex,omitempty"`
	NaturalPairIndices []int                 `json:"naturalPairIndices,omitempty"`
	MeldGroups         [][]int               `json:"meldGroups,omitempty"`
	Config             *HandAndFootWebConfig `json:"config,omitempty"`
}

// HandAndFootWebConfig ハンドアンドフットWeb設定
type HandAndFootWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// HandAndFootWebOutputPlayer ハンドアンドフットWebアウトプットプレイヤー
type HandAndFootWebOutputPlayer struct {
	ID              int              `json:"id"`
	Team            int              `json:"team"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	FootCount       int              `json:"footCount"`
	InFoot          bool             `json:"inFoot"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
}

// HandAndFootWebOutputMeld メルドのアウトプット
type HandAndFootWebOutputMeld struct {
	Cards     []*WebOutputCard `json:"cards"`
	IsNatural bool             `json:"isNatural"`
	IsCanasta bool             `json:"isCanasta"`
	Rank      int              `json:"rank"`
}

// HandAndFootWebOutputTeam チームのアウトプット
type HandAndFootWebOutputTeam struct {
	Team      int                         `json:"team"`
	Melds     []*HandAndFootWebOutputMeld `json:"melds"`
	Red3Count int                         `json:"red3Count"`
	Red3s     []*WebOutputCard            `json:"red3s"`
}

// HandAndFootWebOutput ハンドアンドフットWebアウトプット
type HandAndFootWebOutput struct {
	Players          []*HandAndFootWebOutputPlayer `json:"players"`
	Teams            []*HandAndFootWebOutputTeam   `json:"teams"`
	Phase            int                           `json:"phase"`
	RoundNumber      int                           `json:"roundNumber"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard                `json:"discardTop"`
	DrawPileCount    int                           `json:"drawPileCount"`
	DiscardPileCount int                           `json:"discardPileCount"`
	IsFrozen         bool                          `json:"isFrozen"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerTeam       int                           `json:"winnerTeam"`
	WebOutputBase
	Config HandAndFootWebOutputConfig `json:"config"`
}

// HandAndFootWebOutputConfig ハンドアンドフット設定アウトプット
type HandAndFootWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a HandAndFootConfig from the nested web config, applying bounds checking.
func (c *HandAndFootWebConfig) ToConfig() domain.HandAndFootConfig {
	cfg := domain.DefaultHandAndFootConfig()
	cfg.CpuDifficulty = domain.HandAndFootCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.HandAndFootCpuDifficultyEasy), int(domain.HandAndFootCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 100000)
	return cfg
}

// ToConfig builds a HandAndFootConfig from the web input.
func (p HandAndFootWebInput) ToConfig() domain.HandAndFootConfig {
	return configOrDefault(p.Config, (*HandAndFootWebConfig).ToConfig, domain.DefaultHandAndFootConfig())
}

// HandAndFootWebController ハンドアンドフットWebコントローラークラス
type HandAndFootWebController = GameWebController[usecase.HandAndFootInteractorIF, HandAndFootWebInput, *HandAndFootWebOutput]

// NewHandAndFootWebController and NewHandAndFootWebControllerWithProvider are
// the standard and provider-backed constructors for HandAndFootWebController.
var NewHandAndFootWebController, NewHandAndFootWebControllerWithProvider = webControllerPair[usecase.HandAndFootInteractorIF, HandAndFootWebInput, *HandAndFootWebOutput](
	newHandAndFootDefaultOutput, handAndFootDispatch,
)

func newHandAndFootDefaultOutput(msg string) *HandAndFootWebOutput {
	return &HandAndFootWebOutput{
		Players:       make([]*HandAndFootWebOutputPlayer, 0),
		Teams:         make([]*HandAndFootWebOutputTeam, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func handAndFootDispatch(bc *baseController, w http.ResponseWriter, ci usecase.HandAndFootInteractorIF, param HandAndFootWebInput, newDefault func(string) *HandAndFootWebOutput) bool {
	return dispatchRummyMeld(param.Command, bc, w, rummyMeldFns{
		resetWithConfig: func() string { return ci.ResetWithConfig(param.ToConfig()) },
		drawFromStock:   ci.DrawFromStock,
		drawFromDiscard: func() string { return ci.DrawFromDiscard(param.NaturalPairIndices) },
		meld:            func() string { return ci.Meld(param.MeldGroups) },
		skipMeld:        ci.SkipMeld,
		discard:         ci.Discard,
		goOut:           ci.GoOut,
		nextRound:       ci.NextRound,
		actionLog:       ci.ActionLog,
	}, param.CardIndex, newDefault)
}
