//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ZhengWebConfig 設定 (入力・出力共用)
type ZhengWebConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts ZhengWebConfig to domain.ZhengConfig.
func (c ZhengWebConfig) ToConfig() domain.ZhengConfig {
	return domain.ZhengConfig{
		CpuDifficulty: domain.ZhengCpuDifficulty(c.CpuDifficulty),
	}
}

// ZhengWebInput 争上游 Webインプット
type ZhengWebInput struct {
	BaseWebInput
	Indices []int           `json:"indices"`
	Config  *ZhengWebConfig `json:"config"`
}

// ZhengWebOutputPlayer 争上游 Webアウトプットプレイヤー
type ZhengWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	Rank       int              `json:"rank"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// ZhengWebOutputAction プレイヤー行動記録
type ZhengWebOutputAction struct {
	PlayerIdx   int              `json:"playerIdx"`
	PlayedCards []*WebOutputCard `json:"playedCards"`
}

// ZhengWebOutput 争上游 Webアウトプット
type ZhengWebOutput struct {
	Players           []*ZhengWebOutputPlayer `json:"players"`
	CurrentTurn       int                     `json:"currentTurn"`
	TableCards        []*WebOutputCard        `json:"tableCards"`
	TablePlayType     int                     `json:"tablePlayType"`
	LastPlayPlayerIdx int                     `json:"lastPlayPlayerIdx"`
	GameEndFlag       bool                    `json:"gameEndFlag"`
	CpuActions        []*ZhengWebOutputAction `json:"cpuActions"`
	HumanAction       *ZhengWebOutputAction   `json:"humanAction"`
	Config            ZhengWebConfig          `json:"config"`
	WebOutputBase
}

// ZhengWebController 争上游 Webコントローラークラス
type ZhengWebController = GameWebController[usecase.ZhengInteractorIF, ZhengWebInput, *ZhengWebOutput]

// NewZhengWebController and NewZhengWebControllerWithProvider are
// the standard and provider-backed constructors for ZhengWebController.
var NewZhengWebController, NewZhengWebControllerWithProvider = webControllerPair[usecase.ZhengInteractorIF, ZhengWebInput, *ZhengWebOutput](
	newZhengDefaultOutput, zhengDispatch,
)

func newZhengDefaultOutput(msg string) *ZhengWebOutput {
	return &ZhengWebOutput{
		Players:           make([]*ZhengWebOutputPlayer, 0),
		TableCards:        make([]*WebOutputCard, 0),
		CpuActions:        make([]*ZhengWebOutputAction, 0),
		WebOutputBase:     WebOutputBase{Message: msg},
		LastPlayPlayerIdx: -1,
	}
}

func zhengDispatch(bc *baseController, w http.ResponseWriter, zi usecase.ZhengInteractorIF, param ZhengWebInput, _ func(string) *ZhengWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, zi.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, zi.Reset())
		}
	case "p", "play":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, zi.Play(indices))
	default:
		return dispatchLog(param.Command, bc, w, zi.ActionLog)
	}
	return true
}
