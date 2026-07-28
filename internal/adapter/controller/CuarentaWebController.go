//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CuarentaWebConfig ローカルルール設定 (入力・出力共用)。
type CuarentaWebConfig struct {
	TargetScore   int `json:"targetScore"`
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts CuarentaWebConfig to domain.CuarentaConfig.
func (c CuarentaWebConfig) ToConfig() domain.CuarentaConfig {
	return domain.CuarentaConfig{
		TargetScore:   c.TargetScore,
		CpuDifficulty: domain.CuarentaCpuDifficulty(c.CpuDifficulty),
	}
}

// CuarentaWebInput クアレンタ Web インプット。
type CuarentaWebInput struct {
	BaseWebInput
	HandIndex int                `json:"handIndex"`
	Config    *CuarentaWebConfig `json:"config"`
}

// CuarentaWebOutputPlayer クアレンタ Web アウトプットプレイヤー。
type CuarentaWebOutputPlayer struct {
	ID            int              `json:"id"`
	Team          int              `json:"team"`
	IsHuman       bool             `json:"isHuman"`
	CardCount     int              `json:"cardCount"`
	Cards         []*WebOutputCard `json:"cards"`
	CapturedCount int              `json:"capturedCount"`
}

// CuarentaWebOutputAction 行動記録。
type CuarentaWebOutputAction struct {
	PlayerIdx     int              `json:"playerIdx"`
	PlayedCard    *WebOutputCard   `json:"playedCard"`
	CapturedCards []*WebOutputCard `json:"capturedCards"`
	IsCaida       bool             `json:"isCaida"`
	IsLimpia      bool             `json:"isLimpia"`
	RondaBonus    int              `json:"rondaBonus"`
}

// CuarentaWebOutputScoreDetail 得点内訳。
type CuarentaWebOutputScoreDetail struct {
	CapturedCount map[int]int `json:"capturedCount"`
	Caida         map[int]int `json:"caida"`
	Ronda         map[int]int `json:"ronda"`
	Limpia        map[int]int `json:"limpia"`
	MostCards     int         `json:"mostCards"`
	Gained        map[int]int `json:"gained"`
}

// CuarentaWebOutput クアレンタ Web アウトプット。
type CuarentaWebOutput struct {
	Players         []*CuarentaWebOutputPlayer    `json:"players"`
	CurrentTurn     int                           `json:"currentTurn"`
	TableCards      []*WebOutputCard              `json:"tableCards"`
	LastCaptureIdx  int                           `json:"lastCaptureIdx"`
	GameEndFlag     bool                          `json:"gameEndFlag"`
	Phase           int                           `json:"phase"`
	Config          CuarentaWebConfig             `json:"config"`
	TeamScores      []int                         `json:"teamScores"`
	CpuActions      []*CuarentaWebOutputAction    `json:"cpuActions"`
	HumanAction     *CuarentaWebOutputAction      `json:"humanAction"`
	RemainingDeck   int                           `json:"remainingDeck"`
	RoundWinners    []int                         `json:"roundWinners"`
	LastRoundDetail *CuarentaWebOutputScoreDetail `json:"lastRoundDetail"`
	WebOutputBase
}

// CuarentaWebController クアレンタ Web コントローラークラス。
type CuarentaWebController = GameWebController[usecase.CuarentaInteractorIF, CuarentaWebInput, *CuarentaWebOutput]

// NewCuarentaWebController, NewCuarentaWebControllerWithProvider are the standard
// and provider-backed constructors for CuarentaWebController.
var NewCuarentaWebController, NewCuarentaWebControllerWithProvider = webControllerPair[usecase.CuarentaInteractorIF, CuarentaWebInput, *CuarentaWebOutput](
	newCuarentaDefaultOutput, cuarentaDispatch,
)

func newCuarentaDefaultOutput(msg string) *CuarentaWebOutput {
	return &CuarentaWebOutput{
		Players:       make([]*CuarentaWebOutputPlayer, 0),
		TableCards:    make([]*WebOutputCard, 0),
		TeamScores:    make([]int, 0),
		CpuActions:    make([]*CuarentaWebOutputAction, 0),
		RoundWinners:  make([]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func cuarentaDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CuarentaInteractorIF, param CuarentaWebInput, _ func(string) *CuarentaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, ci.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, ci.Reset())
		}
	case "n", "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "p", "play":
		bc.writePresenterResponse(w, ci.Play(param.HandIndex))
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
