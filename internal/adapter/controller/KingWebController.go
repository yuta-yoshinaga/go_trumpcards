//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KingWebConfig はキングのローカルルール設定 (入力・出力共用)。
type KingWebConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts KingWebConfig to domain.KingConfig.
func (c KingWebConfig) ToConfig() domain.KingConfig {
	return domain.KingConfig{
		CpuDifficulty: domain.KingCpuDifficulty(c.CpuDifficulty),
	}
}

// KingWebInput はキング Web インプット。
type KingWebInput struct {
	BaseWebInput
	Contract  int            `json:"contract"`
	TrumpSuit int            `json:"trumpSuit"`
	HandIndex int            `json:"handIndex"`
	Config    *KingWebConfig `json:"config"`
}

// KingWebOutputPlayer はキング Web アウトプットプレイヤー。
type KingWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	TotalScore int              `json:"totalScore"`
}

// KingWebOutputDealDetail は 1 ディールの得点内訳。
type KingWebOutputDealDetail struct {
	Contract  int         `json:"contract"`
	TrumpSuit int         `json:"trumpSuit"`
	DealerIdx int         `json:"dealerIdx"`
	Gained    map[int]int `json:"gained"`
}

// KingWebOutput はキング Web アウトプット。
type KingWebOutput struct {
	Players         []*KingWebOutputPlayer   `json:"players"`
	Phase           string                   `json:"phase"`
	DealNumber      int                      `json:"dealNumber"`
	TotalDeals      int                      `json:"totalDeals"`
	DealerIdx       int                      `json:"dealerIdx"`
	CurrentTurn     int                      `json:"currentTurn"`
	CurrentContract int                      `json:"currentContract"`
	TrumpSuit       int                      `json:"trumpSuit"`
	TrickNumber     int                      `json:"trickNumber"`
	CurrentTrick    []*WebOutputTrickCard    `json:"currentTrick"`
	LastTrick       []*WebOutputTrickCard    `json:"lastTrick"`
	LastTrickWinner int                      `json:"lastTrickWinner"`
	UsedContracts   []bool                   `json:"usedContracts"`
	PlayableIndices []int                    `json:"playableIndices"`
	GameEndFlag     bool                     `json:"gameEndFlag"`
	Config          KingWebConfig            `json:"config"`
	RoundWinners    []int                    `json:"roundWinners"`
	LastDealDetail  *KingWebOutputDealDetail `json:"lastDealDetail"`
	IsHumanTurn     bool                     `json:"isHumanTurn"`
	Hint            *WebOutputCardHint       `json:"hint,omitempty"`
	WebOutputBase
}

// KingWebController はキング Web コントローラークラス。
type KingWebController = GameWebController[usecase.KingInteractorIF, KingWebInput, *KingWebOutput]

// NewKingWebController, NewKingWebControllerWithProvider are the standard and
// provider-backed constructors for KingWebController.
var NewKingWebController, NewKingWebControllerWithProvider = webControllerPair[usecase.KingInteractorIF, KingWebInput, *KingWebOutput](
	newKingDefaultOutput, kingDispatch,
)

func newKingDefaultOutput(msg string) *KingWebOutput {
	return &KingWebOutput{
		Players:         make([]*KingWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrick:       make([]*WebOutputTrickCard, 0),
		UsedContracts:   make([]bool, 0),
		PlayableIndices: make([]int, 0),
		RoundWinners:    make([]int, 0),
		TotalDeals:      domain.KingTotalDeals,
		LastTrickWinner: -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func kingDispatch(bc *baseController, w http.ResponseWriter, ki usecase.KingInteractorIF, param KingWebInput, newDefault func(string) *KingWebOutput) bool {
	_ = newDefault
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, ki.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, ki.Reset())
		}
	case "n", "next":
		bc.writePresenterResponse(w, ki.NextDeal())
	case "c", "contract":
		bc.writePresenterResponse(w, ki.SelectContract(param.Contract, param.TrumpSuit))
	case "p", "play":
		bc.writePresenterResponse(w, ki.Play(param.HandIndex))
	default:
		return dispatchHintAndLog(param.Command, bc, w, ki.Hint, ki.ActionLog)
	}
	return true
}
