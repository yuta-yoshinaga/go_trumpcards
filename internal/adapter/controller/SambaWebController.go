//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SambaWebInput サンバWebインプット
type SambaWebInput struct {
	BaseWebInput
	CardIndex          *int            `json:"cardIndex,omitempty"`
	NaturalPairIndices []int           `json:"naturalPairIndices,omitempty"`
	MeldGroups         [][]int         `json:"meldGroups,omitempty"`
	Config             *SambaWebConfig `json:"config,omitempty"`
}

// SambaWebConfig サンバWeb設定
type SambaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// SambaWebOutputPlayer サンバWebアウトプットプレイヤー
type SambaWebOutputPlayer struct {
	ID              int                   `json:"id"`
	Team            int                   `json:"team"`
	IsHuman         bool                  `json:"isHuman"`
	CardCount       int                   `json:"cardCount"`
	Cards           []*WebOutputCard      `json:"cards"`
	Melds           []*SambaWebOutputMeld `json:"melds"`
	Red3Count       int                   `json:"red3Count"`
	Red3s           []*WebOutputCard      `json:"red3s"`
	RoundScore      int                   `json:"roundScore"`
	CumulativeScore int                   `json:"cumulativeScore"`
	HasCanasta      bool                  `json:"hasCanasta"`
	HasSamba        bool                  `json:"hasSamba"`
	HasInitMeld     bool                  `json:"hasInitMeld"`
}

// SambaWebOutputMeld メルドのアウトプット。kind でセット/シーケンスを区別し、
// isCanasta/isSamba で完成状態を示す。
type SambaWebOutputMeld struct {
	Cards     []*WebOutputCard `json:"cards"`
	Kind      int              `json:"kind"` // 0 = set, 1 = sequence
	IsNatural bool             `json:"isNatural"`
	IsCanasta bool             `json:"isCanasta"`
	IsSamba   bool             `json:"isSamba"`
	Rank      int              `json:"rank"`
}

// SambaWebOutput サンバWebアウトプット
type SambaWebOutput struct {
	Players          []*SambaWebOutputPlayer `json:"players"`
	TeamScores       []int                   `json:"teamScores"`
	Phase            int                     `json:"phase"`
	RoundNumber      int                     `json:"roundNumber"`
	CurrentPlayerIdx int                     `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard          `json:"discardTop"`
	DrawPileCount    int                     `json:"drawPileCount"`
	DiscardPileCount int                     `json:"discardPileCount"`
	IsFrozen         bool                    `json:"isFrozen"`
	GameEndFlag      bool                    `json:"gameEndFlag"`
	WinnerIdx        int                     `json:"winnerIdx"`
	WebOutputBase
	Config SambaWebOutputConfig `json:"config"`
}

// SambaWebOutputConfig サンバ設定アウトプット
type SambaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a SambaConfig from the nested web config, applying bounds checking.
func (c *SambaWebConfig) ToConfig() domain.SambaConfig {
	cfg := domain.DefaultSambaConfig()
	cfg.CpuDifficulty = domain.SambaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.SambaCpuDifficultyEasy), int(domain.SambaCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 100000)
	return cfg
}

// ToConfig builds a SambaConfig from the web input.
func (p SambaWebInput) ToConfig() domain.SambaConfig {
	return configOrDefault(p.Config, (*SambaWebConfig).ToConfig, domain.DefaultSambaConfig())
}

// SambaWebController サンバWebコントローラークラス
type SambaWebController = GameWebController[usecase.SambaInteractorIF, SambaWebInput, *SambaWebOutput]

// NewSambaWebController and NewSambaWebControllerWithProvider are
// the standard and provider-backed constructors for SambaWebController.
var NewSambaWebController, NewSambaWebControllerWithProvider = webControllerPair[usecase.SambaInteractorIF, SambaWebInput, *SambaWebOutput](
	newSambaDefaultOutput, sambaDispatch,
)

func newSambaDefaultOutput(msg string) *SambaWebOutput {
	return &SambaWebOutput{
		Players:       make([]*SambaWebOutputPlayer, 0),
		TeamScores:    make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func sambaDispatch(bc *baseController, w http.ResponseWriter, ci usecase.SambaInteractorIF, param SambaWebInput, newDefault func(string) *SambaWebOutput) bool {
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
