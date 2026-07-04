//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CariocaWebInput カリオカ Web インプット
type CariocaWebInput struct {
	BaseWebInput
	CardIndex       *int              `json:"cardIndex,omitempty"`
	CardIndices     []int             `json:"cardIndices,omitempty"`
	IndicesPerSlot  [][]int           `json:"indicesPerSlot,omitempty"`
	TargetPlayerIdx *int              `json:"targetPlayerIdx,omitempty"`
	MeldIdx         *int              `json:"meldIdx,omitempty"`
	Config          *CariocaWebConfig `json:"config,omitempty"`
}

// CariocaWebConfig カリオカ Web 設定
type CariocaWebConfig struct {
	PlayerCount         *int `json:"playerCount,omitempty"`
	CpuDifficulty       *int `json:"cpuDifficulty,omitempty"`
	FailContractPenalty *int `json:"failContractPenalty,omitempty"`
}

// CariocaWebOutputMeld メルドのアウトプット
type CariocaWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// CariocaWebOutputPlayer プレイヤーのアウトプット
type CariocaWebOutputPlayer struct {
	ID              int                     `json:"id"`
	IsHuman         bool                    `json:"isHuman"`
	CardCount       int                     `json:"cardCount"`
	Cards           []*WebOutputCard        `json:"cards"`
	Melds           []*CariocaWebOutputMeld `json:"melds"`
	ContractMet     bool                    `json:"contractMet"`
	RoundScore      int                     `json:"roundScore"`
	CumulativeScore int                     `json:"cumulativeScore"`
}

// CariocaWebOutputContractSlot コントラクトスロットのアウトプット
type CariocaWebOutputContractSlot struct {
	Kind int `json:"kind"` // 0=set, 1=run
	Size int `json:"size"`
}

// CariocaWebOutput カリオカ Web アウトプット
type CariocaWebOutput struct {
	Players          []*CariocaWebOutputPlayer       `json:"players"`
	Phase            int                             `json:"phase"`
	RoundNumber      int                             `json:"roundNumber"`
	TotalRounds      int                             `json:"totalRounds"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard                  `json:"discardTop"`
	DrawPileCount    int                             `json:"drawPileCount"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerIdx        int                             `json:"winnerIdx"`
	RoundWinnerIdx   int                             `json:"roundWinnerIdx"`
	ContractSlots    []*CariocaWebOutputContractSlot `json:"contractSlots"`
	WebOutputBase
	Config CariocaWebOutputConfig `json:"config"`
}

// CariocaWebOutputConfig 設定アウトプット
type CariocaWebOutputConfig struct {
	PlayerCount         int `json:"playerCount"`
	CpuDifficulty       int `json:"cpuDifficulty"`
	FailContractPenalty int `json:"failContractPenalty"`
}

// ToConfig builds a CariocaConfig from the nested web config, applying bounds checking.
func (c *CariocaWebConfig) ToConfig() domain.CariocaConfig {
	cfg := domain.DefaultCariocaConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.CariocaPlayerCountMin, domain.CariocaPlayerCountMax)
	cfg.CpuDifficulty = domain.CariocaCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.CariocaCpuDifficultyEasy),
		int(domain.CariocaCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.FailContractPenalty, c.FailContractPenalty, 0, 1000)
	return cfg
}

// ToConfig builds a CariocaConfig from the web input.
func (p CariocaWebInput) ToConfig() domain.CariocaConfig {
	return configOrDefault(p.Config, (*CariocaWebConfig).ToConfig, domain.DefaultCariocaConfig())
}

// CariocaWebController カリオカ Web コントローラー
type CariocaWebController = GameWebController[usecase.CariocaInteractorIF, CariocaWebInput, *CariocaWebOutput]

// NewCariocaWebController / NewCariocaWebControllerWithProvider: 標準／provider 背後の 2 種類のコンストラクタ
var NewCariocaWebController, NewCariocaWebControllerWithProvider = webControllerPair[usecase.CariocaInteractorIF, CariocaWebInput, *CariocaWebOutput](
	newCariocaDefaultOutput, cariocaDispatch,
)

func newCariocaDefaultOutput(msg string) *CariocaWebOutput {
	return &CariocaWebOutput{
		Players:        make([]*CariocaWebOutputPlayer, 0),
		ContractSlots:  make([]*CariocaWebOutputContractSlot, 0),
		WinnerIdx:      -1,
		RoundWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func cariocaDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CariocaInteractorIF, param CariocaWebInput, newDefault func(string) *CariocaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, ci.DrawFromDiscard())
	case "mc", "meldcontract":
		bc.writePresenterResponse(w, ci.MeldContract(param.IndicesPerSlot))
	case "me", "meldextra":
		bc.writePresenterResponse(w, ci.MeldExtra(param.CardIndices))
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
