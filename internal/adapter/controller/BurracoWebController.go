//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BurracoWebInput ブラーコWebインプット
type BurracoWebInput struct {
	BaseWebInput
	CardIndex          *int              `json:"cardIndex,omitempty"`
	NaturalPairIndices []int             `json:"naturalPairIndices,omitempty"`
	MeldGroups         [][]int           `json:"meldGroups,omitempty"`
	Config             *BurracoWebConfig `json:"config,omitempty"`
}

// BurracoWebConfig ブラーコWeb設定
type BurracoWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// BurracoWebOutputPlayer ブラーコWebアウトプットプレイヤー
type BurracoWebOutputPlayer struct {
	ID              int                     `json:"id"`
	IsHuman         bool                    `json:"isHuman"`
	CardCount       int                     `json:"cardCount"`
	Cards           []*WebOutputCard        `json:"cards"`
	Melds           []*BurracoWebOutputMeld `json:"melds"`
	Red3Count       int                     `json:"red3Count"`
	Red3s           []*WebOutputCard        `json:"red3s"`
	RoundScore      int                     `json:"roundScore"`
	CumulativeScore int                     `json:"cumulativeScore"`
	HasBurraco      bool                    `json:"hasBurraco"`
	HasInitMeld     bool                    `json:"hasInitMeld"`
	TookPozzetto    bool                    `json:"tookPozzetto"`
}

// BurracoWebOutputMeld メルドのアウトプット
type BurracoWebOutputMeld struct {
	Cards     []*WebOutputCard `json:"cards"`
	IsNatural bool             `json:"isNatural"`
	IsBurraco bool             `json:"isBurraco"`
	Rank      int              `json:"rank"`
}

// BurracoWebOutput ブラーコWebアウトプット
type BurracoWebOutput struct {
	Players          []*BurracoWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard            `json:"discardTop"`
	DiscardPile      []*WebOutputCard          `json:"discardPile"`
	DrawPileCount    int                       `json:"drawPileCount"`
	DiscardPileCount int                       `json:"discardPileCount"`
	PozzettoCount    int                       `json:"pozzettoCount"`
	IsFrozen         bool                      `json:"isFrozen"`
	GameEndFlag      bool                      `json:"gameEndFlag"`
	WinnerIdx        int                       `json:"winnerIdx"`
	WebOutputBase
	Config BurracoWebOutputConfig `json:"config"`
}

// BurracoWebOutputConfig ブラーコ設定アウトプット
type BurracoWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a BurracoConfig from the nested web config, applying bounds checking.
func (c *BurracoWebConfig) ToConfig() domain.BurracoConfig {
	cfg := domain.DefaultBurracoConfig()
	cfg.CpuDifficulty = domain.BurracoCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.BurracoCpuDifficultyEasy), int(domain.BurracoCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 100000)
	return cfg
}

// ToConfig builds a BurracoConfig from the web input.
func (p BurracoWebInput) ToConfig() domain.BurracoConfig {
	return configOrDefault(p.Config, (*BurracoWebConfig).ToConfig, domain.DefaultBurracoConfig())
}

// BurracoWebController ブラーコWebコントローラークラス
type BurracoWebController = GameWebController[usecase.BurracoInteractorIF, BurracoWebInput, *BurracoWebOutput]

// NewBurracoWebController and NewBurracoWebControllerWithProvider are
// the standard and provider-backed constructors for BurracoWebController.
var NewBurracoWebController, NewBurracoWebControllerWithProvider = webControllerPair[usecase.BurracoInteractorIF, BurracoWebInput, *BurracoWebOutput](
	newBurracoDefaultOutput, burracoDispatch,
)

func newBurracoDefaultOutput(msg string) *BurracoWebOutput {
	return &BurracoWebOutput{
		Players:       make([]*BurracoWebOutputPlayer, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func burracoDispatch(bc *baseController, w http.ResponseWriter, ci usecase.BurracoInteractorIF, param BurracoWebInput, newDefault func(string) *BurracoWebOutput) bool {
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
		hint:            ci.Hint,
	}, param.CardIndex, newDefault)
}
