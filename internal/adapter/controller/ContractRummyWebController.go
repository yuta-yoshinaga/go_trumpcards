//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ContractRummyWebInput コントラクトラミー Web インプット
type ContractRummyWebInput struct {
	BaseWebInput
	CardIndex       *int                    `json:"cardIndex,omitempty"`
	CardIndices     []int                   `json:"cardIndices,omitempty"`
	IndicesPerSlot  [][]int                 `json:"indicesPerSlot,omitempty"`
	TargetPlayerIdx *int                    `json:"targetPlayerIdx,omitempty"`
	MeldIdx         *int                    `json:"meldIdx,omitempty"`
	Config          *ContractRummyWebConfig `json:"config,omitempty"`
}

// ContractRummyWebConfig コントラクトラミー Web 設定
type ContractRummyWebConfig struct {
	CpuDifficulty       *int `json:"cpuDifficulty,omitempty"`
	FailContractPenalty *int `json:"failContractPenalty,omitempty"`
}

// ContractRummyWebOutputMeld メルドのアウトプット
type ContractRummyWebOutputMeld struct {
	Cards []*WebOutputCard `json:"cards"`
}

// ContractRummyWebOutputPlayer プレイヤーのアウトプット
type ContractRummyWebOutputPlayer struct {
	ID              int                           `json:"id"`
	IsHuman         bool                          `json:"isHuman"`
	CardCount       int                           `json:"cardCount"`
	Cards           []*WebOutputCard              `json:"cards"`
	Melds           []*ContractRummyWebOutputMeld `json:"melds"`
	ContractMet     bool                          `json:"contractMet"`
	RoundScore      int                           `json:"roundScore"`
	CumulativeScore int                           `json:"cumulativeScore"`
}

// ContractRummyWebOutputContractSlot コントラクトスロットのアウトプット
type ContractRummyWebOutputContractSlot struct {
	Kind int `json:"kind"` // 0=set, 1=run
	Size int `json:"size"`
}

// ContractRummyWebOutput コントラクトラミー Web アウトプット
type ContractRummyWebOutput struct {
	Players          []*ContractRummyWebOutputPlayer       `json:"players"`
	Phase            int                                   `json:"phase"`
	RoundNumber      int                                   `json:"roundNumber"`
	TotalRounds      int                                   `json:"totalRounds"`
	CurrentPlayerIdx int                                   `json:"currentPlayerIdx"`
	DiscardTop       *WebOutputCard                        `json:"discardTop"`
	DrawPileCount    int                                   `json:"drawPileCount"`
	GameEndFlag      bool                                  `json:"gameEndFlag"`
	WinnerIdx        int                                   `json:"winnerIdx"`
	RoundWinnerIdx   int                                   `json:"roundWinnerIdx"`
	ContractSlots    []*ContractRummyWebOutputContractSlot `json:"contractSlots"`
	WebOutputBase
	Config ContractRummyWebOutputConfig `json:"config"`
}

// ContractRummyWebOutputConfig 設定アウトプット
type ContractRummyWebOutputConfig struct {
	CpuDifficulty       int `json:"cpuDifficulty"`
	FailContractPenalty int `json:"failContractPenalty"`
}

// ToConfig builds a ContractRummyConfig from the nested web config, applying bounds checking.
func (c *ContractRummyWebConfig) ToConfig() domain.ContractRummyConfig {
	cfg := domain.DefaultContractRummyConfig()
	cfg.CpuDifficulty = domain.ContractRummyCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.ContractRummyCpuDifficultyEasy),
		int(domain.ContractRummyCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.FailContractPenalty, c.FailContractPenalty, 0, 1000)
	return cfg
}

// ToConfig builds a ContractRummyConfig from the web input.
func (p ContractRummyWebInput) ToConfig() domain.ContractRummyConfig {
	return configOrDefault(p.Config, (*ContractRummyWebConfig).ToConfig, domain.DefaultContractRummyConfig())
}

// ContractRummyWebController コントラクトラミー Web コントローラー
type ContractRummyWebController = GameWebController[usecase.ContractRummyInteractorIF, ContractRummyWebInput, *ContractRummyWebOutput]

// NewContractRummyWebController / NewContractRummyWebControllerWithProvider: 標準／provider 背後の 2 種類のコンストラクタ
var NewContractRummyWebController, NewContractRummyWebControllerWithProvider = webControllerPair[usecase.ContractRummyInteractorIF, ContractRummyWebInput, *ContractRummyWebOutput](
	newContractRummyDefaultOutput, contractRummyDispatch,
)

func newContractRummyDefaultOutput(msg string) *ContractRummyWebOutput {
	return &ContractRummyWebOutput{
		Players:        make([]*ContractRummyWebOutputPlayer, 0),
		ContractSlots:  make([]*ContractRummyWebOutputContractSlot, 0),
		WinnerIdx:      -1,
		RoundWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func contractRummyDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ContractRummyInteractorIF, param ContractRummyWebInput, newDefault func(string) *ContractRummyWebOutput) bool {
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
