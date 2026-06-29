//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KalookiWebInput カルーキ Web インプット
type KalookiWebInput struct {
	BaseWebInput
	CardIndex       *int              `json:"cardIndex,omitempty"`
	MeldGroups      [][]int           `json:"meldGroups,omitempty"`
	TargetPlayerIdx *int              `json:"targetPlayerIdx,omitempty"`
	MeldIdx         *int              `json:"meldIdx,omitempty"`
	Config          *KalookiWebConfig `json:"config,omitempty"`
}

// KalookiWebConfig カルーキ Web 設定
type KalookiWebConfig struct {
	CpuDifficulty    *int `json:"cpuDifficulty,omitempty"`
	PlayerCount      *int `json:"playerCount,omitempty"`
	OpeningThreshold *int `json:"openingThreshold,omitempty"`
}

// KalookiWebOutputMeld メルドのアウトプット
type KalookiWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// KalookiWebOutputPlayer プレイヤーのアウトプット
type KalookiWebOutputPlayer struct {
	ID              int                     `json:"id"`
	IsHuman         bool                    `json:"isHuman"`
	CardCount       int                     `json:"cardCount"`
	Cards           []*WebOutputCard        `json:"cards"`
	Melds           []*KalookiWebOutputMeld `json:"melds"`
	HasOpened       bool                    `json:"hasOpened"`
	RoundScore      int                     `json:"roundScore"`
	CumulativeScore int                     `json:"cumulativeScore"`
}

// KalookiWebOutput カルーキ Web アウトプット
type KalookiWebOutput struct {
	Players          []*KalookiWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	OpeningThreshold int                       `json:"openingThreshold"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard            `json:"discardTop"`
	DrawPileCount    int                       `json:"drawPileCount"`
	GameEndFlag      bool                      `json:"gameEndFlag"`
	WinnerIdx        int                       `json:"winnerIdx"`
	RoundWinnerIdx   int                       `json:"roundWinnerIdx"`
	WebOutputBase
	Config KalookiWebOutputConfig `json:"config"`
}

// KalookiWebOutputConfig 設定アウトプット
type KalookiWebOutputConfig struct {
	CpuDifficulty    int `json:"cpuDifficulty"`
	PlayerCount      int `json:"playerCount"`
	OpeningThreshold int `json:"openingThreshold"`
}

// ToConfig builds a KalookiConfig from the nested web config, applying bounds checking.
func (c *KalookiWebConfig) ToConfig() domain.KalookiConfig {
	cfg := domain.DefaultKalookiConfig()
	cfg.CpuDifficulty = domain.KalookiCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.KalookiCpuDifficultyEasy),
		int(domain.KalookiCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.KalookiMinPlayers, domain.KalookiMaxPlayers)
	webutil.ApplyBoundedInt(&cfg.OpeningThreshold, c.OpeningThreshold, 0, 1000)
	return cfg
}

// ToConfig builds a KalookiConfig from the web input.
func (p KalookiWebInput) ToConfig() domain.KalookiConfig {
	return configOrDefault(p.Config, (*KalookiWebConfig).ToConfig, domain.DefaultKalookiConfig())
}

// KalookiWebController カルーキ Web コントローラー
type KalookiWebController = GameWebController[usecase.KalookiInteractorIF, KalookiWebInput, *KalookiWebOutput]

// NewKalookiWebController / NewKalookiWebControllerWithProvider: 標準／provider 背後の 2 種類のコンストラクタ
var NewKalookiWebController, NewKalookiWebControllerWithProvider = webControllerPair[usecase.KalookiInteractorIF, KalookiWebInput, *KalookiWebOutput](
	newKalookiDefaultOutput, kalookiDispatch,
)

func newKalookiDefaultOutput(msg string) *KalookiWebOutput {
	return &KalookiWebOutput{
		Players:        make([]*KalookiWebOutputPlayer, 0),
		WinnerIdx:      -1,
		RoundWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func kalookiDispatch(bc *baseController, w http.ResponseWriter, ci usecase.KalookiInteractorIF, param KalookiWebInput, newDefault func(string) *KalookiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, ci.DrawFromDiscard())
	case "m", "meld":
		bc.writePresenterResponse(w, ci.Meld(param.MeldGroups))
	case "lo", "layoff":
		if !requireParam(bc, w, newDefault, param.TargetPlayerIdx == nil || param.MeldIdx == nil || param.CardIndex == nil, "param error: targetPlayerIdx, meldIdx, cardIndex are required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Layoff(*param.TargetPlayerIdx, *param.MeldIdx, *param.CardIndex))
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
