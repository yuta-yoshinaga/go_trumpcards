//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BeziqueWebInput ベジークWebインプット
type BeziqueWebInput struct {
	BaseWebInput
	CardIndex *int              `json:"cardIndex,omitempty"`
	MeldIndex *int              `json:"meldIndex,omitempty"`
	Config    *BeziqueWebConfig `json:"config,omitempty"`
}

// BeziqueWebConfig ベジークWeb設定
type BeziqueWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// BeziqueWebOutputPlayer ベジークWebアウトプットプレイヤー
type BeziqueWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// BeziqueWebOutputMeld 宣言可能な役
type BeziqueWebOutputMeld struct {
	Type   int `json:"type"`
	Suit   int `json:"suit"`
	Points int `json:"points"`
}

// BeziqueWebOutputHint ヒント出力
type BeziqueWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	MeldIndex *int   `json:"meldIndex,omitempty"`
	Reason    string `json:"reason"`
}

// BeziqueWebOutput ベジークWebアウトプット
type BeziqueWebOutput struct {
	Players          []*BeziqueWebOutputPlayer `json:"players"`
	DealPoints       []int                     `json:"dealPoints"`
	DealMeldPoints   []int                     `json:"dealMeldPoints"`
	MatchScore       []int                     `json:"matchScore"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	TrickNumber      int                       `json:"trickNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                       `json:"leadPlayerIdx"`
	DealerIdx        int                       `json:"dealerIdx"`
	TrumpSuit        int                       `json:"trumpSuit"`
	TrumpCard        *WebOutputCard            `json:"trumpCard,omitempty"`
	CurrentTrick     []*WebOutputTrickCard     `json:"currentTrick"`
	StockRemaining   int                       `json:"stockRemaining"`
	IsEndgame        bool                      `json:"isEndgame"`
	AvailableMelds   []*BeziqueWebOutputMeld   `json:"availableMelds"`
	GameEndFlag      bool                      `json:"gameEndFlag"`
	WinnerIdx        int                       `json:"winnerIdx"`
	Hint             *BeziqueWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config BeziqueWebOutputConfig `json:"config"`
}

// BeziqueWebOutputConfig ベジーク設定アウトプット
type BeziqueWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds a BeziqueConfig from the nested web config, applying bounds checking.
func (c *BeziqueWebConfig) ToConfig() domain.BeziqueConfig {
	cfg := domain.DefaultBeziqueConfig()
	cfg.CpuDifficulty = domain.BeziqueCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.BeziqueCpuDifficultyEasy), int(domain.BeziqueCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	cfg.TargetScore = webutil.BoundedIntPtr(c.TargetScore, 100, domain.BeziqueMaxTargetScore, cfg.TargetScore)
	return cfg
}

// ToConfig builds a BeziqueConfig from the web input.
func (p BeziqueWebInput) ToConfig() domain.BeziqueConfig {
	return configOrDefault(p.Config, (*BeziqueWebConfig).ToConfig, domain.DefaultBeziqueConfig())
}

// BeziqueWebController ベジークWebコントローラークラス
type BeziqueWebController = GameWebController[usecase.BeziqueInteractorIF, BeziqueWebInput, *BeziqueWebOutput]

// NewBeziqueWebController and NewBeziqueWebControllerWithProvider are
// the standard and provider-backed constructors for BeziqueWebController.
var NewBeziqueWebController, NewBeziqueWebControllerWithProvider = webControllerPair[usecase.BeziqueInteractorIF, BeziqueWebInput, *BeziqueWebOutput](
	newBeziqueDefaultOutput, beziqueDispatch,
)

func newBeziqueDefaultOutput(msg string) *BeziqueWebOutput {
	return &BeziqueWebOutput{
		Players:        make([]*BeziqueWebOutputPlayer, 0),
		DealPoints:     make([]int, 0),
		DealMeldPoints: make([]int, 0),
		MatchScore:     make([]int, 0),
		CurrentTrick:   make([]*WebOutputTrickCard, 0),
		AvailableMelds: make([]*BeziqueWebOutputMeld, 0),
		WinnerIdx:      -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func beziqueDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BeziqueInteractorIF, param BeziqueWebInput, newDefault func(string) *BeziqueWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(*param.CardIndex))
	case "m", "meld":
		if !requireParam(bc, w, newDefault, param.MeldIndex == nil, "param error: meldIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.DeclareMeld(*param.MeldIndex))
	case "s", "skip":
		bc.writePresenterResponse(w, bi.SkipMeld())
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
