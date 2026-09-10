//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BiribaWebInput ビリバWebインプット
type BiribaWebInput struct {
	BaseWebInput
	CardIndex          *int             `json:"cardIndex,omitempty"`
	NaturalPairIndices []int            `json:"naturalPairIndices,omitempty"`
	MeldGroups         [][]int          `json:"meldGroups,omitempty"`
	Config             *BiribaWebConfig `json:"config,omitempty"`
}

// BiribaWebConfig ビリバWeb設定
type BiribaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// BiribaWebOutputPlayer ビリバWebアウトプットプレイヤー
type BiribaWebOutputPlayer struct {
	ID              int                    `json:"id"`
	IsHuman         bool                   `json:"isHuman"`
	CardCount       int                    `json:"cardCount"`
	Cards           []*WebOutputCard       `json:"cards"`
	Melds           []*BiribaWebOutputMeld `json:"melds"`
	Red3Count       int                    `json:"red3Count"`
	Red3s           []*WebOutputCard       `json:"red3s"`
	RoundScore      int                    `json:"roundScore"`
	CumulativeScore int                    `json:"cumulativeScore"`
	HasBiriba       bool                   `json:"hasBiriba"`
	HasInitMeld     bool                   `json:"hasInitMeld"`
	TookPozzetto    bool                   `json:"tookPozzetto"`
}

// BiribaWebOutputMeld メルドのアウトプット
type BiribaWebOutputMeld struct {
	Cards     []*WebOutputCard `json:"cards"`
	IsNatural bool             `json:"isNatural"`
	IsBiriba  bool             `json:"isBiriba"`
	Rank      int              `json:"rank"`
}

// BiribaWebOutput ビリバWebアウトプット
type BiribaWebOutput struct {
	Players          []*BiribaWebOutputPlayer `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard           `json:"discardTop"`
	DiscardPile      []*WebOutputCard         `json:"discardPile"`
	DrawPileCount    int                      `json:"drawPileCount"`
	DiscardPileCount int                      `json:"discardPileCount"`
	PozzettoCount    int                      `json:"pozzettoCount"`
	IsFrozen         bool                     `json:"isFrozen"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerIdx        int                      `json:"winnerIdx"`
	// Hint は人間の手番のときだけ入る。空オブジェクトを返すと「行動できない」と
	// 読めるので、無いときは省略する。
	Hint *BiribaWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config BiribaWebOutputConfig `json:"config"`
}

// BiribaWebOutputHint ヒントのアウトプット。
//
// **インデックスで運ぶ。**カードそのものを送ると、フロントが手札の並びを
// 変えたときに別の札を指す。CUI が使うドメインのヒントと同じ値なので、
// 2 つの画面が同じ盤面で違う手を勧めることがなくなる (#5628)。
type BiribaWebOutputHint struct {
	// Action は "draw_stock" / "draw_discard" / "meld" / "skip_meld" / "discard"。
	Action string `json:"action"`
	// Indices は対象カードの手札インデックス (draw_stock / skip_meld では空)。
	Indices []int `json:"indices,omitempty"`
	// Reason は理由の i18n キー (接頭辞なし)。
	Reason string `json:"reason"`
}

// BiribaWebOutputConfig ビリバ設定アウトプット
type BiribaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a BiribaConfig from the nested web config, applying bounds checking.
func (c *BiribaWebConfig) ToConfig() domain.BiribaConfig {
	cfg := domain.DefaultBiribaConfig()
	cfg.CpuDifficulty = domain.BiribaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.BiribaCpuDifficultyEasy), int(domain.BiribaCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 100000)
	return cfg
}

// ToConfig builds a BiribaConfig from the web input.
func (p BiribaWebInput) ToConfig() domain.BiribaConfig {
	return configOrDefault(p.Config, (*BiribaWebConfig).ToConfig, domain.DefaultBiribaConfig())
}

// BiribaWebController ビリバWebコントローラークラス
type BiribaWebController = GameWebController[usecase.BiribaInteractorIF, BiribaWebInput, *BiribaWebOutput]

// NewBiribaWebController and NewBiribaWebControllerWithProvider are
// the standard and provider-backed constructors for BiribaWebController.
var NewBiribaWebController, NewBiribaWebControllerWithProvider = webControllerPair[usecase.BiribaInteractorIF, BiribaWebInput, *BiribaWebOutput](
	newBiribaDefaultOutput, biribaDispatch,
)

func newBiribaDefaultOutput(msg string) *BiribaWebOutput {
	return &BiribaWebOutput{
		Players:       make([]*BiribaWebOutputPlayer, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func biribaDispatch(bc *baseController, w http.ResponseWriter, ci usecase.BiribaInteractorIF, param BiribaWebInput, newDefault func(string) *BiribaWebOutput) bool {
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
