//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CanastaWebInput カナスタWebインプット
type CanastaWebInput struct {
	BaseWebInput
	CardIndex          *int              `json:"cardIndex,omitempty"`
	NaturalPairIndices []int             `json:"naturalPairIndices,omitempty"`
	MeldGroups         [][]int           `json:"meldGroups,omitempty"`
	Config             *CanastaWebConfig `json:"config,omitempty"`
}

// CanastaWebConfig カナスタWeb設定
type CanastaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// CanastaWebOutputPlayer カナスタWebアウトプットプレイヤー
type CanastaWebOutputPlayer struct {
	ID              int                     `json:"id"`
	IsHuman         bool                    `json:"isHuman"`
	CardCount       int                     `json:"cardCount"`
	Cards           []*WebOutputCard        `json:"cards"`
	Melds           []*CanastaWebOutputMeld `json:"melds"`
	Red3Count       int                     `json:"red3Count"`
	Red3s           []*WebOutputCard        `json:"red3s"`
	RoundScore      int                     `json:"roundScore"`
	CumulativeScore int                     `json:"cumulativeScore"`
	HasCanasta      bool                    `json:"hasCanasta"`
	HasInitMeld     bool                    `json:"hasInitMeld"`
}

// CanastaWebOutputMeld メルドのアウトプット
type CanastaWebOutputMeld struct {
	Cards     []*WebOutputCard `json:"cards"`
	IsNatural bool             `json:"isNatural"`
	IsCanasta bool             `json:"isCanasta"`
	Rank      int              `json:"rank"`
}

// CanastaWebOutput カナスタWebアウトプット
type CanastaWebOutput struct {
	Players          []*CanastaWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard            `json:"discardTop"`
	DrawPileCount    int                       `json:"drawPileCount"`
	DiscardPileCount int                       `json:"discardPileCount"`
	IsFrozen         bool                      `json:"isFrozen"`
	GameEndFlag      bool                      `json:"gameEndFlag"`
	WinnerIdx        int                       `json:"winnerIdx"`
	WebOutputBase
	Config CanastaWebOutputConfig `json:"config"`
}

// CanastaWebOutputConfig カナスタ設定アウトプット
type CanastaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a CanastaConfig from the nested web config, applying bounds checking.
func (c *CanastaWebConfig) ToConfig() domain.CanastaConfig {
	cfg := domain.DefaultCanastaConfig()
	cfg.CpuDifficulty = domain.CanastaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.CanastaCpuDifficultyEasy), int(domain.CanastaCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 100000)
	return cfg
}

// ToConfig builds a CanastaConfig from the web input.
func (p CanastaWebInput) ToConfig() domain.CanastaConfig {
	return configOrDefault(p.Config, (*CanastaWebConfig).ToConfig, domain.DefaultCanastaConfig())
}

// CanastaWebController カナスタWebコントローラークラス
type CanastaWebController = GameWebController[usecase.CanastaInteractorIF, CanastaWebInput, *CanastaWebOutput]

// NewCanastaWebController and NewCanastaWebControllerWithProvider are
// the standard and provider-backed constructors for CanastaWebController.
var NewCanastaWebController, NewCanastaWebControllerWithProvider = webControllerPair[usecase.CanastaInteractorIF, CanastaWebInput, *CanastaWebOutput](
	newCanastaDefaultOutput, canastaDispatch,
)

func newCanastaDefaultOutput(msg string) *CanastaWebOutput {
	return &CanastaWebOutput{
		Players:       make([]*CanastaWebOutputPlayer, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func canastaDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CanastaInteractorIF, param CanastaWebInput, newDefault func(string) *CanastaWebOutput) bool {
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
