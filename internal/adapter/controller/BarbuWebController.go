//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BarbuWebConfig ローカルルール設定 (入力・出力共用)
type BarbuWebConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts BarbuWebConfig to domain.BarbuConfig.
func (c BarbuWebConfig) ToConfig() domain.BarbuConfig {
	return domain.BarbuConfig{
		CpuDifficulty: domain.BarbuCpuDifficulty(c.CpuDifficulty),
	}
}

// BarbuWebInput バルブ Web インプット
type BarbuWebInput struct {
	BaseWebInput
	Contract     int             `json:"contract"`
	TrumpSuit    int             `json:"trumpSuit"`
	HandIndex    int             `json:"handIndex"`
	TableIndices []int           `json:"tableIndices"`
	Config       *BarbuWebConfig `json:"config"`
}

// BarbuWebOutputPlayer バルブ Web アウトプットプレイヤー
type BarbuWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	DominoRank int              `json:"dominoRank"`
	TotalScore int              `json:"totalScore"`
}

// BarbuWebOutputTrickCard トリック中の 1 枚
type BarbuWebOutputTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// BarbuWebOutputDealDetail 1 ディールの得点内訳
type BarbuWebOutputDealDetail struct {
	Contract  int         `json:"contract"`
	TrumpSuit int         `json:"trumpSuit"`
	DealerIdx int         `json:"dealerIdx"`
	Gained    map[int]int `json:"gained"`
}

// BarbuWebOutput バルブ Web アウトプット
type BarbuWebOutput struct {
	Players         []*BarbuWebOutputPlayer     `json:"players"`
	Phase           string                      `json:"phase"`
	DealNumber      int                         `json:"dealNumber"`
	TotalDeals      int                         `json:"totalDeals"`
	DealerIdx       int                         `json:"dealerIdx"`
	CurrentTurn     int                         `json:"currentTurn"`
	CurrentContract int                         `json:"currentContract"`
	TrumpSuit       int                         `json:"trumpSuit"`
	TrickNumber     int                         `json:"trickNumber"`
	CurrentTrick    []*BarbuWebOutputTrickCard  `json:"currentTrick"`
	LastTrick       []*BarbuWebOutputTrickCard  `json:"lastTrick"`
	LastTrickWinner int                         `json:"lastTrickWinner"`
	TablePlaced     []int                       `json:"tablePlaced"`
	DominoPlayable  []int                       `json:"dominoPlayable"`
	UsedContracts   []bool                      `json:"usedContracts"`
	GameEndFlag     bool                        `json:"gameEndFlag"`
	Config          BarbuWebConfig              `json:"config"`
	RoundWinners    []int                       `json:"roundWinners"`
	LastDealDetail  *BarbuWebOutputDealDetail   `json:"lastDealDetail"`
	DealHistory     []*BarbuWebOutputDealDetail `json:"dealHistory"`
	WebOutputBase
}

// BarbuWebController バルブ Web コントローラークラス
type BarbuWebController = GameWebController[usecase.BarbuInteractorIF, BarbuWebInput, *BarbuWebOutput]

// NewBarbuWebController, NewBarbuWebControllerWithProvider are the standard and
// provider-backed constructors for BarbuWebController.
var NewBarbuWebController, NewBarbuWebControllerWithProvider = webControllerPair[usecase.BarbuInteractorIF, BarbuWebInput, *BarbuWebOutput](
	newBarbuDefaultOutput, barbuDispatch,
)

func newBarbuDefaultOutput(msg string) *BarbuWebOutput {
	return &BarbuWebOutput{
		Players:        make([]*BarbuWebOutputPlayer, 0),
		CurrentTrick:   make([]*BarbuWebOutputTrickCard, 0),
		LastTrick:      make([]*BarbuWebOutputTrickCard, 0),
		TablePlaced:    make([]int, 0),
		DominoPlayable: make([]int, 0),
		UsedContracts:  make([]bool, 0),
		RoundWinners:   make([]int, 0),
		DealHistory:    make([]*BarbuWebOutputDealDetail, 0),
		TotalDeals:     domain.BarbuTotalDeals,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

// barbuMaxIndices caps client-supplied table-index slices for the play command.
const barbuMaxIndices = 13

func barbuDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BarbuInteractorIF, param BarbuWebInput, defaultOutput func(string) *BarbuWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, bi.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, bi.Reset())
		}
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextDeal())
	case "c", "contract":
		bc.writePresenterResponse(w, bi.SelectContract(param.Contract, param.TrumpSuit))
	case "p", "play":
		if len(param.TableIndices) > barbuMaxIndices {
			bc.writeJsonResponse(w, http.StatusBadRequest, defaultOutput("too many indices"))
			return true
		}
		bc.writePresenterResponse(w, bi.Play(param.HandIndex, param.TableIndices))
	default:
		return dispatchLog(param.Command, bc, w, bi.ActionLog)
	}
	return true
}
