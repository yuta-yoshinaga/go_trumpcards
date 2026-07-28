//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SevenBridgeWebInput セブンブリッジ Web インプット
type SevenBridgeWebInput struct {
	BaseWebInput
	CardIndex       *int                  `json:"cardIndex,omitempty"`
	CardIndices     []int                 `json:"cardIndices,omitempty"`
	TargetPlayerIdx *int                  `json:"targetPlayerIdx,omitempty"`
	MeldIdx         *int                  `json:"meldIdx,omitempty"`
	Config          *SevenBridgeWebConfig `json:"config,omitempty"`
}

// SevenBridgeWebConfig セブンブリッジ Web 設定
type SevenBridgeWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// SevenBridgeWebOutputPlayer セブンブリッジ Web アウトプットプレイヤー
type SevenBridgeWebOutputPlayer struct {
	ID              int                         `json:"id"`
	IsHuman         bool                        `json:"isHuman"`
	CardCount       int                         `json:"cardCount"`
	Cards           []*WebOutputCard            `json:"cards"`
	Melds           []*SevenBridgeWebOutputMeld `json:"melds"`
	RoundScore      int                         `json:"roundScore"`
	CumulativeScore int                         `json:"cumulativeScore"`
}

// SevenBridgeWebOutputMeld メルドのアウトプット
type SevenBridgeWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// SevenBridgeWebOutput セブンブリッジ Web アウトプット
type SevenBridgeWebOutput struct {
	Players          []*SevenBridgeWebOutputPlayer `json:"players"`
	Phase            int                           `json:"phase"`
	RoundNumber      int                           `json:"roundNumber"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard                `json:"discardTop"`
	DrawPileCount    int                           `json:"drawPileCount"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerIdx        int                           `json:"winnerIdx"`
	RoundWinnerIdx   int                           `json:"roundWinnerIdx"`
	WebOutputBase
	Config SevenBridgeWebOutputConfig `json:"config"`
}

// SevenBridgeWebOutputConfig セブンブリッジ設定アウトプット
type SevenBridgeWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a SevenBridgeConfig from the nested web config, applying bounds checking.
func (c *SevenBridgeWebConfig) ToConfig() domain.SevenBridgeConfig {
	cfg := domain.DefaultSevenBridgeConfig()
	cfg.CpuDifficulty = domain.SevenBridgeCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.SevenBridgeCpuDifficultyEasy), int(domain.SevenBridgeCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a SevenBridgeConfig from the web input.
func (p SevenBridgeWebInput) ToConfig() domain.SevenBridgeConfig {
	return configOrDefault(p.Config, (*SevenBridgeWebConfig).ToConfig, domain.DefaultSevenBridgeConfig())
}

// SevenBridgeWebController セブンブリッジ Web コントローラー
type SevenBridgeWebController = GameWebController[usecase.SevenBridgeInteractorIF, SevenBridgeWebInput, *SevenBridgeWebOutput]

// NewSevenBridgeWebController / NewSevenBridgeWebControllerWithProvider: 標準／provider 背後の 2 種類のコンストラクタ
var NewSevenBridgeWebController, NewSevenBridgeWebControllerWithProvider = webControllerPair[usecase.SevenBridgeInteractorIF, SevenBridgeWebInput, *SevenBridgeWebOutput](
	newSevenBridgeDefaultOutput, sevenBridgeDispatch,
)

func newSevenBridgeDefaultOutput(msg string) *SevenBridgeWebOutput {
	return &SevenBridgeWebOutput{
		Players:        make([]*SevenBridgeWebOutputPlayer, 0),
		WinnerIdx:      -1,
		RoundWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func sevenBridgeDispatch(bc *baseController, w http.ResponseWriter, ci usecase.SevenBridgeInteractorIF, param SevenBridgeWebInput, newDefault func(string) *SevenBridgeWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "pon":
		bc.writePresenterResponse(w, ci.ClaimPon(param.CardIndices))
	case "chi":
		bc.writePresenterResponse(w, ci.ClaimChi(param.CardIndices))
	case "m", "meld":
		bc.writePresenterResponse(w, ci.Meld(param.CardIndices))
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
