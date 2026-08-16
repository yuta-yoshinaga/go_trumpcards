//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ScopaWebConfig ローカルルール設定 (入力・出力共用)
type ScopaWebConfig struct {
	TargetScore   int `json:"targetScore"`
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts ScopaWebConfig to domain.ScopaConfig.
func (c ScopaWebConfig) ToConfig() domain.ScopaConfig {
	return domain.ScopaConfig{
		TargetScore:   c.TargetScore,
		CpuDifficulty: domain.ScopaCpuDifficulty(c.CpuDifficulty),
	}
}

// ScopaWebInput スコパWebインプット
type ScopaWebInput struct {
	BaseWebInput
	HandIndex    int             `json:"handIndex"`
	TableIndices []int           `json:"tableIndices"`
	Config       *ScopaWebConfig `json:"config"`
}

// ScopaWebOutputPlayer スコパWebアウトプットプレイヤー
type ScopaWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	CardCount     int              `json:"cardCount"`
	Cards         []*WebOutputCard `json:"cards"`
	CapturedCount int              `json:"capturedCount"`
	ScopaCount    int              `json:"scopaCount"`
	TotalScore    int              `json:"totalScore"`
}

// ScopaWebOutputAction 行動記録
type ScopaWebOutputAction struct {
	PlayerIdx     int              `json:"playerIdx"`
	PlayedCard    *WebOutputCard   `json:"playedCard"`
	CapturedCards []*WebOutputCard `json:"capturedCards"`
	IsScopa       bool             `json:"isScopa"`
}

// ScopaWebOutputScoreDetail 得点内訳
type ScopaWebOutputScoreDetail struct {
	Cards         map[int]int `json:"cards"`
	Diamonds      map[int]int `json:"diamonds"`
	Sevens        map[int]int `json:"sevens"`
	HasSetteBello int         `json:"hasSetteBello"`
	Scopas        map[int]int `json:"scopas"`
	Gained        map[int]int `json:"gained"`
}

// ScopaWebOutput スコパWebアウトプット
type ScopaWebOutput struct {
	Players         []*ScopaWebOutputPlayer    `json:"players"`
	CurrentTurn     int                        `json:"currentTurn"`
	TableCards      []*WebOutputCard           `json:"tableCards"`
	LastCaptureIdx  int                        `json:"lastCaptureIdx"`
	GameEndFlag     bool                       `json:"gameEndFlag"`
	Phase           string                     `json:"phase"`
	Config          ScopaWebConfig             `json:"config"`
	CpuActions      []*ScopaWebOutputAction    `json:"cpuActions"`
	HumanAction     *ScopaWebOutputAction      `json:"humanAction"`
	RemainingDeck   int                        `json:"remainingDeck"`
	PacksDealt      int                        `json:"packsDealt"`
	RoundWinners    []int                      `json:"roundWinners"`
	LastRoundDetail *ScopaWebOutputScoreDetail `json:"lastRoundDetail"`
	WebOutputBase
}

// ScopaWebController スコパWebコントローラークラス
type ScopaWebController = GameWebController[usecase.ScopaInteractorIF, ScopaWebInput, *ScopaWebOutput]

// NewScopaWebController, NewScopaWebControllerWithProvider are the standard and
// provider-backed constructors for ScopaWebController.
var NewScopaWebController, NewScopaWebControllerWithProvider = webControllerPair[usecase.ScopaInteractorIF, ScopaWebInput, *ScopaWebOutput](
	newScopaDefaultOutput, scopaDispatch,
)

func newScopaDefaultOutput(msg string) *ScopaWebOutput {
	return &ScopaWebOutput{
		Players:       make([]*ScopaWebOutputPlayer, 0),
		TableCards:    make([]*WebOutputCard, 0),
		CpuActions:    make([]*ScopaWebOutputAction, 0),
		RoundWinners:  make([]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

// scopaMaxIndices caps client-supplied table-index slices for the play command.
const scopaMaxIndices = 40

func scopaDispatch(bc *baseController, w http.ResponseWriter, si usecase.ScopaInteractorIF, param ScopaWebInput, defaultOutput func(string) *ScopaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, si.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, si.Reset())
		}
	case "n", "next":
		bc.writePresenterResponse(w, si.NextRound())
	case "p", "play":
		if len(param.TableIndices) > scopaMaxIndices {
			bc.writeJsonResponse(w, http.StatusBadRequest, defaultOutput("too many indices"))
			return true
		}
		bc.writePresenterResponse(w, si.Play(param.HandIndex, param.TableIndices))
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
