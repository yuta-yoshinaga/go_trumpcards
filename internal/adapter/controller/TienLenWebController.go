//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TienLenWebConfig 設定 (入力・出力共用)
type TienLenWebConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts TienLenWebConfig to domain.TienLenConfig.
func (c TienLenWebConfig) ToConfig() domain.TienLenConfig {
	return domain.TienLenConfig{
		CpuDifficulty: domain.TienLenCpuDifficulty(c.CpuDifficulty),
	}
}

// TienLenWebInput Tien Len Webインプット
type TienLenWebInput struct {
	BaseWebInput
	Indices []int             `json:"indices"`
	Config  *TienLenWebConfig `json:"config"`
}

// TienLenWebOutputPlayer Tien Len Webアウトプットプレイヤー
type TienLenWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	Rank       int              `json:"rank"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// TienLenWebOutputAction プレイヤー行動記録
type TienLenWebOutputAction struct {
	PlayerIdx   int              `json:"playerIdx"`
	PlayedCards []*WebOutputCard `json:"playedCards"`
}

// TienLenWebOutput Tien Len Webアウトプット
type TienLenWebOutput struct {
	Players           []*TienLenWebOutputPlayer `json:"players"`
	CurrentTurn       int                       `json:"currentTurn"`
	TableCards        []*WebOutputCard          `json:"tableCards"`
	TablePlayType     int                       `json:"tablePlayType"`
	LastPlayPlayerIdx int                       `json:"lastPlayPlayerIdx"`
	GameEndFlag       bool                      `json:"gameEndFlag"`
	CpuActions        []*TienLenWebOutputAction `json:"cpuActions"`
	HumanAction       *TienLenWebOutputAction   `json:"humanAction"`
	Config            TienLenWebConfig          `json:"config"`
	WebOutputBase
}

// TienLenWebController Tien Len Webコントローラークラス
type TienLenWebController = GameWebController[usecase.TienLenInteractorIF, TienLenWebInput, *TienLenWebOutput]

// NewTienLenWebController and NewTienLenWebControllerWithProvider are
// the standard and provider-backed constructors for TienLenWebController.
var NewTienLenWebController, NewTienLenWebControllerWithProvider = webControllerPair[usecase.TienLenInteractorIF, TienLenWebInput, *TienLenWebOutput](
	newTienLenDefaultOutput, tienLenDispatch,
)

func newTienLenDefaultOutput(msg string) *TienLenWebOutput {
	return &TienLenWebOutput{
		Players:           make([]*TienLenWebOutputPlayer, 0),
		TableCards:        make([]*WebOutputCard, 0),
		CpuActions:        make([]*TienLenWebOutputAction, 0),
		WebOutputBase:     WebOutputBase{Message: msg},
		LastPlayPlayerIdx: -1,
	}
}

func tienLenDispatch(bc *baseController, w http.ResponseWriter, tli usecase.TienLenInteractorIF, param TienLenWebInput, _ func(string) *TienLenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, tli.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, tli.Reset())
		}
	case "p", "play":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, tli.Play(indices))
	default:
		return dispatchHintAndLog(param.Command, bc, w, tli.Hint, tli.ActionLog)
	}
	return true
}
