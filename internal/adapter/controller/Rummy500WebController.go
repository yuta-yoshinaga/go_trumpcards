//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// Rummy500WebInput Rummy 500Webインプット
type Rummy500WebInput struct {
	BaseWebInput
	CardIndex   *int               `json:"cardIndex,omitempty"`
	CardIndices []int              `json:"cardIndices,omitempty"`
	DiscardIdx  *int               `json:"discardIdx,omitempty"`
	MeldOwner   *int               `json:"meldOwner,omitempty"`
	MeldIdx     *int               `json:"meldIdx,omitempty"`
	Config      *Rummy500WebConfig `json:"config,omitempty"`
}

// Rummy500WebConfig Rummy 500Web設定
type Rummy500WebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// Rummy500WebOutputPlayer Rummy 500Webアウトプットプレイヤー
type Rummy500WebOutputPlayer struct {
	ID              int                      `json:"id"`
	IsHuman         bool                     `json:"isHuman"`
	CardCount       int                      `json:"cardCount"`
	Cards           []*WebOutputCard         `json:"cards"`
	RoundScore      int                      `json:"roundScore"`
	CumulativeScore int                      `json:"cumulativeScore"`
	LaidMelds       []*Rummy500WebOutputMeld `json:"laidMelds"`
}

// Rummy500WebOutputMeld メルドのアウトプット
type Rummy500WebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// Rummy500WebOutput Rummy 500Webアウトプット
// Rummy500LayoffTarget は 1 枚のカードを置ける既存メルドの場所。
type Rummy500LayoffTarget struct {
	Owner   int `json:"owner"`
	MeldIdx int `json:"meldIdx"`
}

type Rummy500WebOutput struct {
	Players []*Rummy500WebOutputPlayer `json:"players"`
	// LayoffTargets[i] は人間の手札 i 番目を置けるメルドの場所一覧。
	// 押せるボタンが必ず通るようにするために使う (#4832)。
	LayoffTargets    [][]Rummy500LayoffTarget `json:"layoffTargets"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	DiscardPile      []*WebOutputCard         `json:"discardPile"`
	DrawPileCount    int                      `json:"drawPileCount"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerIdx        int                      `json:"winnerIdx"`
	RoundEnderIdx    int                      `json:"roundEnderIdx"`
	WebOutputBase
	Config Rummy500WebOutputConfig `json:"config"`
}

// Rummy500WebOutputConfig Rummy 500設定アウトプット
type Rummy500WebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a Rummy500Config from the nested web config, applying bounds checking.
func (c *Rummy500WebConfig) ToConfig() domain.Rummy500Config {
	cfg := domain.DefaultRummy500Config()
	cfg.CpuDifficulty = domain.Rummy500CpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.Rummy500CpuDifficultyEasy), int(domain.Rummy500CpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 10000)
	return cfg
}

// ToConfig builds a Rummy500Config from the web input.
func (p Rummy500WebInput) ToConfig() domain.Rummy500Config {
	return configOrDefault(p.Config, (*Rummy500WebConfig).ToConfig, domain.DefaultRummy500Config())
}

// Rummy500WebController Rummy 500Webコントローラークラス
type Rummy500WebController = GameWebController[usecase.Rummy500InteractorIF, Rummy500WebInput, *Rummy500WebOutput]

// NewRummy500WebController and NewRummy500WebControllerWithProvider are
// the standard and provider-backed constructors for Rummy500WebController.
var NewRummy500WebController, NewRummy500WebControllerWithProvider = webControllerPair[usecase.Rummy500InteractorIF, Rummy500WebInput, *Rummy500WebOutput](
	newRummy500DefaultOutput, rummy500Dispatch,
)

func newRummy500DefaultOutput(msg string) *Rummy500WebOutput {
	return &Rummy500WebOutput{
		Players:       make([]*Rummy500WebOutputPlayer, 0),
		WinnerIdx:     -1,
		RoundEnderIdx: -1,
		DiscardPile:   make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func rummy500Dispatch(bc *baseController, w http.ResponseWriter, ci usecase.Rummy500InteractorIF, param Rummy500WebInput, newDefault func(string) *Rummy500WebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "dd", "drawdiscard":
		idx := -1
		if param.DiscardIdx != nil {
			idx = *param.DiscardIdx
		}
		bc.writePresenterResponse(w, ci.DrawFromDiscard(idx))
	case "m", "meld":
		bc.writePresenterResponse(w, ci.Meld(param.CardIndices))
	case "lo", "layoff":
		if !requireParam(bc, w, newDefault, param.MeldOwner == nil || param.MeldIdx == nil || param.CardIndex == nil, "param error: meldOwner, meldIdx and cardIndex are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Layoff(*param.MeldOwner, *param.MeldIdx, *param.CardIndex))
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Discard(*param.CardIndex))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
