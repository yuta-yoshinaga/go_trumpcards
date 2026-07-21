package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PigsTailWebInput ぶたのしっぽWebインプット
type PigsTailWebInput struct {
	BaseWebInput
	CpuHesitationEnabled bool `json:"cpuHesitationEnabled"`
	PlayerCount          *int `json:"playerCount,omitempty"`
}

// PigsTailWebOutputPlayer ぶたのしっぽWebアウトプットプレイヤー
type PigsTailWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
}

// PigsTailWebOutputCpuAction CPUターンの行動記録
type PigsTailWebOutputCpuAction struct {
	DrawPlayerIdx int            `json:"drawPlayerIdx"`
	DrawnCard     *WebOutputCard `json:"drawnCard"`
	PenaltyFlag   bool           `json:"penaltyFlag"`
	PenaltyCount  int            `json:"penaltyCount"`
	HesitationMs  int            `json:"hesitationMs,omitempty"`
}

// PigsTailWebOutput ぶたのしっぽWebアウトプット
type PigsTailWebOutput struct {
	Players      []*PigsTailWebOutputPlayer    `json:"players"`
	CircleCount  int                           `json:"circleCount"`
	CenterTop    *WebOutputCard                `json:"centerTop"`
	CenterCount  int                           `json:"centerCount"`
	CurrentTurn  int                           `json:"currentTurn"`
	GameEndFlag  bool                          `json:"gameEndFlag"`
	LoserIdx     int                           `json:"loserIdx"`
	LastDrawCard *WebOutputCard                `json:"lastDrawCard"`
	LastPenalty  bool                          `json:"lastPenalty"`
	CpuActions   []*PigsTailWebOutputCpuAction `json:"cpuActions"`
	HumanAction  *PigsTailWebOutputCpuAction   `json:"humanAction"`
	WebOutputBase
}

// PigsTailWebController ぶたのしっぽWebコントローラークラス
type PigsTailWebController = GameWebController[usecase.PigsTailInteractorIF, PigsTailWebInput, *PigsTailWebOutput]

// NewPigsTailWebController and NewPigsTailWebControllerWithProvider are
// the standard and provider-backed constructors for PigsTailWebController.
var NewPigsTailWebController, NewPigsTailWebControllerWithProvider = webControllerPair[usecase.PigsTailInteractorIF, PigsTailWebInput, *PigsTailWebOutput](
	newPigsTailDefaultOutput, pigsTailDispatch,
)

func newPigsTailDefaultOutput(msg string) *PigsTailWebOutput {
	return &PigsTailWebOutput{
		Players:       make([]*PigsTailWebOutputPlayer, 0),
		CpuActions:    make([]*PigsTailWebOutputCpuAction, 0),
		LoserIdx:      -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func pigsTailDispatch(bc *baseController, w http.ResponseWriter, pti usecase.PigsTailInteractorIF, param PigsTailWebInput, _ func(string) *PigsTailWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		cfg := domain.DefaultPigsTailConfig()
		cfg.CpuHesitationEnabled = param.CpuHesitationEnabled
		webutil.ApplyBoundedInt(&cfg.PlayerCount, param.PlayerCount, domain.PigsTailMinPlayers, domain.PigsTailMaxPlayers)
		bc.writePresenterResponse(w, pti.Reset(cfg))
	case "d", "draw":
		bc.writePresenterResponse(w, pti.Action(0))
	default:
		return dispatchLog(param.Command, bc, w, pti.ActionLog)
	}
	return true
}
