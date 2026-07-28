//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TichuWebConfig ティチュー設定 (入力・出力共用)
type TichuWebConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts TichuWebConfig to domain.TichuConfig.
func (c TichuWebConfig) ToConfig() domain.TichuConfig {
	return domain.TichuConfig{
		CpuDifficulty: domain.TichuCpuDifficulty(c.CpuDifficulty),
	}
}

// TichuWebInput ティチューWebインプット
type TichuWebInput struct {
	BaseWebInput
	Indices  []int           `json:"indices"`  // play コマンド用
	DeclType *int            `json:"declType"` // declare コマンド用
	Config   *TichuWebConfig `json:"config"`   // リセット時の設定 (省略可)
}

// TichuWebOutputPlayer ティチューWebアウトプットプレイヤー
type TichuWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	Team       int              `json:"team"`
	Rank       int              `json:"rank"`
	DeclType   int              `json:"declType"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// TichuWebOutputAction ティチューのプレイヤー行動記録
type TichuWebOutputAction struct {
	PlayerIdx   int              `json:"playerIdx"`
	PlayedCards []*WebOutputCard `json:"playedCards"`
	DeclType    int              `json:"declType"`
	IsPass      bool             `json:"isPass"`
}

// TichuWebOutput ティチューWebアウトプット
type TichuWebOutput struct {
	Players     []*TichuWebOutputPlayer `json:"players"`
	Phase       string                  `json:"phase"`
	CurrentTurn int                     `json:"currentTurn"`
	TableCards  []*WebOutputCard        `json:"tableCards"`
	TableCombo  string                  `json:"tableCombo"`
	LastPlayIdx int                     `json:"lastPlayIdx"`
	StartLeader int                     `json:"startLeader"`
	FinishOrder []int                   `json:"finishOrder"`
	Scores      [2]int                  `json:"scores"`
	IsOneTwo    bool                    `json:"isOneTwo"`
	BombCount   int                     `json:"bombCount"`
	GameEndFlag bool                    `json:"gameEndFlag"`
	Config      TichuWebConfig          `json:"config"`
	CpuActions  []*TichuWebOutputAction `json:"cpuActions"`
	HumanAction *TichuWebOutputAction   `json:"humanAction"`
	WebOutputBase
}

// TichuWebController ティチューWebコントローラー
type TichuWebController = GameWebController[usecase.TichuInteractorIF, TichuWebInput, *TichuWebOutput]

// NewTichuWebController and NewTichuWebControllerWithProvider are the standard
// and provider-backed constructors for TichuWebController.
var NewTichuWebController, NewTichuWebControllerWithProvider = webControllerPair[usecase.TichuInteractorIF, TichuWebInput, *TichuWebOutput](
	newTichuDefaultOutput, tichuDispatch,
)

func newTichuDefaultOutput(msg string) *TichuWebOutput {
	return &TichuWebOutput{
		Players:       make([]*TichuWebOutputPlayer, 0),
		TableCards:    make([]*WebOutputCard, 0),
		FinishOrder:   make([]int, 0),
		CpuActions:    make([]*TichuWebOutputAction, 0),
		LastPlayIdx:   -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func tichuDispatch(bc *baseController, w http.ResponseWriter, tgi usecase.TichuInteractorIF, param TichuWebInput, _ func(string) *TichuWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, tgi.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, tgi.Reset())
		}
	case "declare":
		value := 0
		if param.DeclType != nil {
			value = *param.DeclType
		}
		bc.writePresenterResponse(w, tgi.Declare(value))
	case "p", "play":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, tgi.Play(indices))
	default:
		return dispatchLog(param.Command, bc, w, tgi.ActionLog)
	}
	return true
}
