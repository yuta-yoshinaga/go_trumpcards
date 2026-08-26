//go:build !js || !wasm || extra4

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"net/http"
)

// DoudizhuWebConfig 斗地主設定 (入力・出力共用)
type DoudizhuWebConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts DoudizhuWebConfig to domain.DoudizhuConfig.
func (c DoudizhuWebConfig) ToConfig() domain.DoudizhuConfig {
	return domain.DoudizhuConfig{
		CpuDifficulty: domain.DoudizhuCpuDifficulty(c.CpuDifficulty),
	}
}

// DoudizhuWebInput 斗地主Webインプット
type DoudizhuWebInput struct {
	BaseWebInput
	Indices  []int              `json:"indices"`  // play コマンド用
	BidValue *int               `json:"bidValue"` // bid コマンド用
	Config   *DoudizhuWebConfig `json:"config"`   // リセット時の設定 (省略可)
}

// DoudizhuWebOutputPlayer 斗地主Webアウトプットプレイヤー
type DoudizhuWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	IsLandlord bool             `json:"isLandlord"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// DoudizhuWebOutputAction 斗地主のプレイヤー行動記録
type DoudizhuWebOutputAction struct {
	PlayerIdx   int              `json:"playerIdx"`
	PlayedCards []*WebOutputCard `json:"playedCards"`
	BidValue    int              `json:"bidValue"`
}

// DoudizhuWebOutput 斗地主Webアウトプット
type DoudizhuWebOutput struct {
	Players     []*DoudizhuWebOutputPlayer    `json:"players"`
	Phase       string                        `json:"phase"`
	CurrentTurn int                           `json:"currentTurn"`
	TableCards  []*WebOutputCard              `json:"tableCards"`
	TableCombo  string                        `json:"tableCombo"`
	KittyCards  []*WebOutputCard              `json:"kittyCards"`
	LandlordIdx int                           `json:"landlordIdx"`
	BaseBid     int                           `json:"baseBid"`
	HighestBid  int                           `json:"highestBid"`
	BombCount   int                           `json:"bombCount"`
	Scores      [domain.DoudizhuPlayerCnt]int `json:"scores"`
	GameEndFlag bool                          `json:"gameEndFlag"`
	Config      DoudizhuWebConfig             `json:"config"`
	CpuActions  []*DoudizhuWebOutputAction    `json:"cpuActions"`
	HumanAction *DoudizhuWebOutputAction      `json:"humanAction"`
	WebOutputBase
}

// DoudizhuWebController 斗地主Webコントローラー
type DoudizhuWebController = GameWebController[usecase.DoudizhuInteractorIF, DoudizhuWebInput, *DoudizhuWebOutput]

// NewDoudizhuWebController and NewDoudizhuWebControllerWithProvider are
// the standard and provider-backed constructors for DoudizhuWebController.
var NewDoudizhuWebController, NewDoudizhuWebControllerWithProvider = webControllerPair[usecase.DoudizhuInteractorIF, DoudizhuWebInput, *DoudizhuWebOutput](
	newDoudizhuDefaultOutput, doudizhuDispatch,
)

func newDoudizhuDefaultOutput(msg string) *DoudizhuWebOutput {
	return &DoudizhuWebOutput{
		Players:       make([]*DoudizhuWebOutputPlayer, 0),
		TableCards:    make([]*WebOutputCard, 0),
		KittyCards:    make([]*WebOutputCard, 0),
		CpuActions:    make([]*DoudizhuWebOutputAction, 0),
		LandlordIdx:   -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func doudizhuDispatch(bc *baseController, w http.ResponseWriter, dgi usecase.DoudizhuInteractorIF, param DoudizhuWebInput, _ func(string) *DoudizhuWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, dgi.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, dgi.Reset())
		}
	case "bid":
		value := 0
		if param.BidValue != nil {
			value = *param.BidValue
		}
		bc.writePresenterResponse(w, dgi.Bid(value))
	case "p", "play":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, dgi.Play(indices))
	default:
		return dispatchLog(param.Command, bc, w, dgi.ActionLog)
	}
	return true
}
