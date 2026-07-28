//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PishtiWebConfig はローカルルール設定 (入力・出力共用)。
type PishtiWebConfig struct {
	PlayerCnt     int `json:"playerCnt"`
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts PishtiWebConfig to domain.PishtiConfig.
func (c PishtiWebConfig) ToConfig() domain.PishtiConfig {
	return domain.PishtiConfig{
		PlayerCnt:     c.PlayerCnt,
		CpuDifficulty: domain.PishtiCpuDifficulty(c.CpuDifficulty),
	}
}

// PishtiWebInput は Pişti Web インプット。
type PishtiWebInput struct {
	BaseWebInput
	HandIndex int              `json:"handIndex"`
	Config    *PishtiWebConfig `json:"config"`
}

// PishtiWebOutputPlayer は Pişti Web アウトプットプレイヤー。
type PishtiWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	CardCount     int              `json:"cardCount"`
	Cards         []*WebOutputCard `json:"cards"`
	CapturedCount int              `json:"capturedCount"`
	PistiBonus    int              `json:"pistiBonus"`
	FinalScore    int              `json:"finalScore"`
}

// PishtiWebOutput は Pişti Web アウトプット。
type PishtiWebOutput struct {
	Players        []*PishtiWebOutputPlayer `json:"players"`
	CurrentTurn    int                      `json:"currentTurn"`
	Pile           []*WebOutputCard         `json:"pile"`
	PileTop        *WebOutputCard           `json:"pileTop"`
	PileCount      int                      `json:"pileCount"`
	LastCaptureIdx int                      `json:"lastCaptureIdx"`
	GameEndFlag    bool                     `json:"gameEndFlag"`
	Phase          string                   `json:"phase"`
	Config         PishtiWebConfig          `json:"config"`
	RemainingDeck  int                      `json:"remainingDeck"`
	Winners        []int                    `json:"winners"`
	FinalScores    []int                    `json:"finalScores"`
	WebOutputBase
}

// PishtiWebController は Pişti Web コントローラークラス。
type PishtiWebController = GameWebController[usecase.PishtiInteractorIF, PishtiWebInput, *PishtiWebOutput]

// NewPishtiWebController, NewPishtiWebControllerWithProvider are the standard
// and provider-backed constructors for PishtiWebController.
var NewPishtiWebController, NewPishtiWebControllerWithProvider = webControllerPair[usecase.PishtiInteractorIF, PishtiWebInput, *PishtiWebOutput](
	newPishtiDefaultOutput, pishtiDispatch,
)

func newPishtiDefaultOutput(msg string) *PishtiWebOutput {
	return &PishtiWebOutput{
		Players:       make([]*PishtiWebOutputPlayer, 0),
		Pile:          make([]*WebOutputCard, 0),
		Winners:       make([]int, 0),
		FinalScores:   make([]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func pishtiDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PishtiInteractorIF, param PishtiWebInput, _ func(string) *PishtiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, pi.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, pi.Reset())
		}
	case "n", "next":
		bc.writePresenterResponse(w, pi.NextRound())
	case "p", "play":
		bc.writePresenterResponse(w, pi.Play(param.HandIndex))
	default:
		return dispatchLog(param.Command, bc, w, pi.ActionLog)
	}
	return true
}
