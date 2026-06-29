//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ThreeThirteenWebInput スリー・サーティーン Web インプット
type ThreeThirteenWebInput struct {
	BaseWebInput
	CardIndex *int                    `json:"cardIndex,omitempty"`
	Config    *ThreeThirteenWebConfig `json:"config,omitempty"`
}

// ThreeThirteenWebConfig スリー・サーティーン Web 設定
type ThreeThirteenWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PlayerCount   *int `json:"playerCount,omitempty"`
}

// ThreeThirteenWebOutputPlayer プレイヤーのアウトプット
type ThreeThirteenWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Deadwood        int              `json:"deadwood"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
}

// ThreeThirteenWebOutput スリー・サーティーン Web アウトプット
type ThreeThirteenWebOutput struct {
	Players          []*ThreeThirteenWebOutputPlayer `json:"players"`
	Phase            int                             `json:"phase"`
	Round            int                             `json:"round"`
	WildRank         int                             `json:"wildRank"`
	DealCount        int                             `json:"dealCount"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	KnockerIdx       int                             `json:"knockerIdx"`
	DiscardTop       *WebOutputCard                  `json:"discardTop"`
	DrawPileCount    int                             `json:"drawPileCount"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerIdx        int                             `json:"winnerIdx"`
	WebOutputBase
	Config ThreeThirteenWebOutputConfig `json:"config"`
}

// ThreeThirteenWebOutputConfig 設定アウトプット
type ThreeThirteenWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PlayerCount   int `json:"playerCount"`
}

// ToConfig builds a ThreeThirteenConfig from the nested web config, applying bounds checking.
func (c *ThreeThirteenWebConfig) ToConfig() domain.ThreeThirteenConfig {
	cfg := domain.DefaultThreeThirteenConfig()
	cfg.CpuDifficulty = domain.ThreeThirteenCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.ThreeThirteenCpuDifficultyEasy),
		int(domain.ThreeThirteenCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.ThreeThirteenMinPlayers, domain.ThreeThirteenMaxPlayers)
	return cfg
}

// ToConfig builds a ThreeThirteenConfig from the web input.
func (p ThreeThirteenWebInput) ToConfig() domain.ThreeThirteenConfig {
	return configOrDefault(p.Config, (*ThreeThirteenWebConfig).ToConfig, domain.DefaultThreeThirteenConfig())
}

// ThreeThirteenWebController スリー・サーティーン Web コントローラー
type ThreeThirteenWebController = GameWebController[usecase.ThreeThirteenInteractorIF, ThreeThirteenWebInput, *ThreeThirteenWebOutput]

// NewThreeThirteenWebController / NewThreeThirteenWebControllerWithProvider: 標準／provider 背後の 2 種類のコンストラクタ
var NewThreeThirteenWebController, NewThreeThirteenWebControllerWithProvider = webControllerPair[usecase.ThreeThirteenInteractorIF, ThreeThirteenWebInput, *ThreeThirteenWebOutput](
	newThreeThirteenDefaultOutput, threeThirteenDispatch,
)

func newThreeThirteenDefaultOutput(msg string) *ThreeThirteenWebOutput {
	return &ThreeThirteenWebOutput{
		Players:       make([]*ThreeThirteenWebOutputPlayer, 0),
		WinnerIdx:     -1,
		KnockerIdx:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func threeThirteenDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ThreeThirteenInteractorIF, param ThreeThirteenWebInput, newDefault func(string) *ThreeThirteenWebOutput) bool {
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
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
