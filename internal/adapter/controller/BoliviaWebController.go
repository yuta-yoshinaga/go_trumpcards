//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BoliviaWebInput ボリビアWebインプット
type BoliviaWebInput struct {
	BaseWebInput
	CardIndex          *int              `json:"cardIndex,omitempty"`
	NaturalPairIndices []int             `json:"naturalPairIndices,omitempty"`
	MeldGroups         [][]int           `json:"meldGroups,omitempty"`
	Config             *BoliviaWebConfig `json:"config,omitempty"`
}

// BoliviaWebConfig ボリビアWeb設定
type BoliviaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// BoliviaWebOutputPlayer ボリビアWebアウトプットプレイヤー
type BoliviaWebOutputPlayer struct {
	ID              int                     `json:"id"`
	Team            int                     `json:"team"`
	IsHuman         bool                    `json:"isHuman"`
	CardCount       int                     `json:"cardCount"`
	Cards           []*WebOutputCard        `json:"cards"`
	Melds           []*BoliviaWebOutputMeld `json:"melds"`
	Red3Count       int                     `json:"red3Count"`
	Red3s           []*WebOutputCard        `json:"red3s"`
	RoundScore      int                     `json:"roundScore"`
	CumulativeScore int                     `json:"cumulativeScore"`
	HasCanasta      bool                    `json:"hasCanasta"`
	// HasEscalera は完成したエスカレラを持っているか。**上がりに要るのはこちら。**
	HasEscalera bool `json:"hasEscalera"`
	// HasBolivia は完成したボリビア (ワイルド 7 枚) を持っているか。点が重いだけ。
	HasBolivia  bool `json:"hasBolivia"`
	HasInitMeld bool `json:"hasInitMeld"`
}

// BoliviaWebOutputMeld メルドのアウトプット。kind でセット/シーケンスを区別し、
// isCanasta/isBolivia で完成状態を示す。
type BoliviaWebOutputMeld struct {
	Cards     []*WebOutputCard `json:"cards"`
	Kind      int              `json:"kind"` // 0 = set, 1 = sequence
	IsNatural bool             `json:"isNatural"`
	IsCanasta bool             `json:"isCanasta"`
	// IsEscalera は 7 枚のエスカレラか (ワイルド無しの同スート連番)。
	IsEscalera bool `json:"isEscalera"`
	// IsBolivia は 7 枚のワイルドメルドか。ゲーム名の由来で 2500 点。
	IsBolivia bool `json:"isBolivia"`
	Rank      int  `json:"rank"`
}

// BoliviaWebOutput ボリビアWebアウトプット
type BoliviaWebOutput struct {
	Players          []*BoliviaWebOutputPlayer `json:"players"`
	TeamScores       []int                     `json:"teamScores"`
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
	Config BoliviaWebOutputConfig `json:"config"`
}

// BoliviaWebOutputConfig ボリビア設定アウトプット
type BoliviaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a BoliviaConfig from the nested web config, applying bounds checking.
func (c *BoliviaWebConfig) ToConfig() domain.BoliviaConfig {
	cfg := domain.DefaultBoliviaConfig()
	cfg.CpuDifficulty = domain.BoliviaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.BoliviaCpuDifficultyEasy), int(domain.BoliviaCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 100000)
	return cfg
}

// ToConfig builds a BoliviaConfig from the web input.
func (p BoliviaWebInput) ToConfig() domain.BoliviaConfig {
	return configOrDefault(p.Config, (*BoliviaWebConfig).ToConfig, domain.DefaultBoliviaConfig())
}

// BoliviaWebController ボリビアWebコントローラークラス
type BoliviaWebController = GameWebController[usecase.BoliviaInteractorIF, BoliviaWebInput, *BoliviaWebOutput]

// NewBoliviaWebController and NewBoliviaWebControllerWithProvider are
// the standard and provider-backed constructors for BoliviaWebController.
var NewBoliviaWebController, NewBoliviaWebControllerWithProvider = webControllerPair[usecase.BoliviaInteractorIF, BoliviaWebInput, *BoliviaWebOutput](
	newBoliviaDefaultOutput, boliviaDispatch,
)

func newBoliviaDefaultOutput(msg string) *BoliviaWebOutput {
	return &BoliviaWebOutput{
		Players:       make([]*BoliviaWebOutputPlayer, 0),
		TeamScores:    make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func boliviaDispatch(bc *baseController, w http.ResponseWriter, ci usecase.BoliviaInteractorIF, param BoliviaWebInput, newDefault func(string) *BoliviaWebOutput) bool {
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
