//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PanWebInput パングインゲ Web インプット
type PanWebInput struct {
	BaseWebInput
	CardIndex   *int          `json:"cardIndex,omitempty"`
	CardIndices []int         `json:"cardIndices,omitempty"`
	MeldOwner   *int          `json:"meldOwner,omitempty"`
	MeldIdx     *int          `json:"meldIdx,omitempty"`
	Config      *PanWebConfig `json:"config,omitempty"`
}

// PanWebConfig パングインゲ Web 設定
type PanWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// PanWebOutputMeld メルドのアウトプット
type PanWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// PanWebOutputPlayer プレイヤーのアウトプット
type PanWebOutputPlayer struct {
	ID              int                 `json:"id"`
	IsHuman         bool                `json:"isHuman"`
	CardCount       int                 `json:"cardCount"`
	Cards           []*WebOutputCard    `json:"cards"`
	LaidMelds       []*PanWebOutputMeld `json:"laidMelds"`
	MeldedCount     int                 `json:"meldedCount"`
	Chips           int                 `json:"chips"`
	HandPoints      int                 `json:"handPoints"`
	RoundScore      int                 `json:"roundScore"`
	CumulativeScore int                 `json:"cumulativeScore"`
}

// PanWebOutput パングインゲ Web アウトプット
type PanWebOutput struct {
	Players          []*PanWebOutputPlayer `json:"players"`
	Phase            int                   `json:"phase"`
	RoundNumber      int                   `json:"roundNumber"`
	TargetRounds     int                   `json:"targetRounds"`
	CurrentPlayerIdx int                   `json:"currentPlayerIdx"`
	DealerIdx        int                   `json:"dealerIdx"`
	DiscardTop       *WebOutputCard        `json:"discardTop"`
	DrawPileCount    int                   `json:"drawPileCount"`
	DeckSize         int                   `json:"deckSize"`
	WinMeldCount     int                   `json:"winMeldCount"`
	GameEndFlag      bool                  `json:"gameEndFlag"`
	WinnerIdx        int                   `json:"winnerIdx"`
	PanDeclarerIdx   int                   `json:"panDeclarerIdx"`
	WebOutputBase
	Config PanWebOutputConfig `json:"config"`
}

// PanWebOutputConfig 設定アウトプット
type PanWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds a PanConfig from the nested web config, applying bounds checking.
func (c *PanWebConfig) ToConfig() domain.PanConfig {
	cfg := domain.DefaultPanConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.PanPlayerCountMin, domain.PanPlayerCountMax)
	cfg.CpuDifficulty = domain.PanCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.PanCpuDifficultyEasy),
		int(domain.PanCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 100)
	return cfg
}

// ToConfig builds a PanConfig from the web input.
func (p PanWebInput) ToConfig() domain.PanConfig {
	return configOrDefault(p.Config, (*PanWebConfig).ToConfig, domain.DefaultPanConfig())
}

// PanWebController パングインゲ Web コントローラー
type PanWebController = GameWebController[usecase.PanInteractorIF, PanWebInput, *PanWebOutput]

// NewPanWebController / NewPanWebControllerWithProvider: 標準／provider 背後の 2 種類のコンストラクタ
var NewPanWebController, NewPanWebControllerWithProvider = webControllerPair[usecase.PanInteractorIF, PanWebInput, *PanWebOutput](
	newPanDefaultOutput, panDispatch,
)

func newPanDefaultOutput(msg string) *PanWebOutput {
	return &PanWebOutput{
		Players:        make([]*PanWebOutputPlayer, 0),
		WinnerIdx:      -1,
		PanDeclarerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func panDispatch(bc *baseController, w http.ResponseWriter, ci usecase.PanInteractorIF, param PanWebInput, newDefault func(string) *PanWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, ci.DrawFromDiscard())
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
