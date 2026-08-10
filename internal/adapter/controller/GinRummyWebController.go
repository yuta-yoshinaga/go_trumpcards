//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GinRummyWebInput ジンラミーWebインプット
type GinRummyWebInput struct {
	BaseWebInput
	CardIndex   *int               `json:"cardIndex,omitempty"`
	CardIndices []int              `json:"cardIndices,omitempty"`
	Config      *GinRummyWebConfig `json:"config,omitempty"`
}

// GinRummyWebConfig ジンラミーWeb設定
type GinRummyWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// GinRummyWebOutputPlayer ジンラミーWebアウトプットプレイヤー
type GinRummyWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
}

// GinRummyWebOutputMeld メルドのアウトプット
type GinRummyWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// GinRummyWebOutput ジンラミーWebアウトプット
type GinRummyWebOutput struct {
	Players          []*GinRummyWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard             `json:"discardTop"`
	DrawPileCount    int                        `json:"drawPileCount"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	KnockerIdx       int                        `json:"knockerIdx"`
	KnockerMelds     []*GinRummyWebOutputMeld   `json:"knockerMelds"`
	// LayoffTargets[i] は人間の手札 i 番目を足せるノッカーのメルド番号一覧。
	// レイオフフェーズで「どれを付け足せるか」を画面が示すために使う (#4823)。
	LayoffTargets   [][]int          `json:"layoffTargets"`
	KnockerDeadwood []*WebOutputCard `json:"knockerDeadwood"`
	IsGin           bool             `json:"isGin"`
	WebOutputBase
	Config GinRummyWebOutputConfig `json:"config"`
}

// GinRummyWebOutputConfig ジンラミー設定アウトプット
type GinRummyWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a GinRummyConfig from the nested web config, applying bounds checking.
func (c *GinRummyWebConfig) ToConfig() domain.GinRummyConfig {
	cfg := domain.DefaultGinRummyConfig()
	cfg.CpuDifficulty = domain.GinRummyCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.GinRummyCpuDifficultyEasy), int(domain.GinRummyCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a GinRummyConfig from the web input.
func (p GinRummyWebInput) ToConfig() domain.GinRummyConfig {
	return configOrDefault(p.Config, (*GinRummyWebConfig).ToConfig, domain.DefaultGinRummyConfig())
}

// GinRummyWebController ジンラミーWebコントローラークラス
type GinRummyWebController = GameWebController[usecase.GinRummyInteractorIF, GinRummyWebInput, *GinRummyWebOutput]

// NewGinRummyWebController and NewGinRummyWebControllerWithProvider are
// the standard and provider-backed constructors for GinRummyWebController.
var NewGinRummyWebController, NewGinRummyWebControllerWithProvider = webControllerPair[usecase.GinRummyInteractorIF, GinRummyWebInput, *GinRummyWebOutput](
	newGinRummyDefaultOutput, ginRummyDispatch,
)

func newGinRummyDefaultOutput(msg string) *GinRummyWebOutput {
	return &GinRummyWebOutput{
		Players:         make([]*GinRummyWebOutputPlayer, 0),
		WinnerIdx:       -1,
		KnockerIdx:      -1,
		KnockerMelds:    make([]*GinRummyWebOutputMeld, 0),
		KnockerDeadwood: make([]*WebOutputCard, 0),
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func ginRummyDispatch(bc *baseController, w http.ResponseWriter, ci usecase.GinRummyInteractorIF, param GinRummyWebInput, newDefault func(string) *GinRummyWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, ci.DrawFromDiscard())
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Discard(*param.CardIndex))
	case "k", "knock":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Knock(*param.CardIndex))
	case "lo", "layoff":
		bc.writePresenterResponse(w, ci.Layoff(param.CardIndices))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
