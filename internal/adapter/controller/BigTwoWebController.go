package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BigTwoWebConfig 設定 (入力・出力共用)
type BigTwoWebConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts BigTwoWebConfig to domain.BigTwoConfig.
func (c BigTwoWebConfig) ToConfig() domain.BigTwoConfig {
	return domain.BigTwoConfig{
		CpuDifficulty: domain.BigTwoCpuDifficulty(c.CpuDifficulty),
	}
}

// BigTwoWebInput Big Two Webインプット
type BigTwoWebInput struct {
	BaseWebInput
	Indices []int            `json:"indices"`
	Config  *BigTwoWebConfig `json:"config"`
}

// BigTwoWebOutputPlayer Big Two Webアウトプットプレイヤー
type BigTwoWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	Rank       int              `json:"rank"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// BigTwoWebOutputAction プレイヤー行動記録
type BigTwoWebOutputAction struct {
	PlayerIdx   int              `json:"playerIdx"`
	PlayedCards []*WebOutputCard `json:"playedCards"`
}

// BigTwoWebOutput Big Two Webアウトプット
type BigTwoWebOutput struct {
	Players           []*BigTwoWebOutputPlayer `json:"players"`
	CurrentTurn       int                      `json:"currentTurn"`
	TableCards        []*WebOutputCard         `json:"tableCards"`
	TablePlayType     int                      `json:"tablePlayType"`
	LastPlayPlayerIdx int                      `json:"lastPlayPlayerIdx"`
	GameEndFlag       bool                     `json:"gameEndFlag"`
	CpuActions        []*BigTwoWebOutputAction `json:"cpuActions"`
	HumanAction       *BigTwoWebOutputAction   `json:"humanAction"`
	Config            BigTwoWebConfig          `json:"config"`
	WebOutputBase
}

// BigTwoWebController Big Two Webコントローラークラス
type BigTwoWebController = GameWebController[usecase.BigTwoInteractorIF, BigTwoWebInput, *BigTwoWebOutput]

// NewBigTwoWebController and NewBigTwoWebControllerWithProvider are
// the standard and provider-backed constructors for BigTwoWebController.
var NewBigTwoWebController, NewBigTwoWebControllerWithProvider = webControllerPair[usecase.BigTwoInteractorIF, BigTwoWebInput, *BigTwoWebOutput](
	newBigTwoDefaultOutput, bigTwoDispatch,
)

func newBigTwoDefaultOutput(msg string) *BigTwoWebOutput {
	return &BigTwoWebOutput{
		Players:           make([]*BigTwoWebOutputPlayer, 0),
		TableCards:        make([]*WebOutputCard, 0),
		CpuActions:        make([]*BigTwoWebOutputAction, 0),
		WebOutputBase:     WebOutputBase{Message: msg},
		LastPlayPlayerIdx: -1,
	}
}

func bigTwoDispatch(bc *baseController, w http.ResponseWriter, bti usecase.BigTwoInteractorIF, param BigTwoWebInput, _ func(string) *BigTwoWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, bti.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, bti.Reset())
		}
	case "p", "play":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, bti.Play(indices))
	default:
		return dispatchLog(param.Command, bc, w, bti.ActionLog)
	}
	return true
}
