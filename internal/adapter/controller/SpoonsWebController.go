//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpoonsWebInput はスプーンの Web 入力。
type SpoonsWebInput struct {
	BaseWebInput
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *SpoonsWebConfig `json:"config,omitempty"`
}

// SpoonsWebConfig はスプーンの設定リクエスト。
type SpoonsWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// SpoonsWebPlayer はプレイヤー出力。
type SpoonsWebPlayer struct {
	Name       string           `json:"name"`
	IsHuman    bool             `json:"isHuman"`
	HandSize   int              `json:"handSize"`
	Hand       []*WebOutputCard `json:"hand"`
	Letters    int              `json:"letters"`
	Eliminated bool             `json:"eliminated"`
	HasSpoon   bool             `json:"hasSpoon"`
}

// SpoonsWebOutput はスプーンの Web 出力。
type SpoonsWebOutput struct {
	Phase            int                `json:"phase"`
	GameEndFlag      bool               `json:"gameEndFlag"`
	WinnerIdx        int                `json:"winnerIdx"`
	CurrentPlayerIdx int                `json:"currentPlayerIdx"`
	FeederIdx        int                `json:"feederIdx"`
	IsHumanTurn      bool               `json:"isHumanTurn"`
	SpoonsRemaining  int                `json:"spoonsRemaining"`
	GrabWindowOpen   bool               `json:"grabWindowOpen"`
	FirstGrabberIdx  int                `json:"firstGrabberIdx"`
	RoundLoserIdx    int                `json:"roundLoserIdx"`
	RoundNumber      int                `json:"roundNumber"`
	DrawPileSize     int                `json:"drawPileSize"`
	Players          []*SpoonsWebPlayer `json:"players"`
	CpuDifficulty    int                `json:"cpuDifficulty"`
	WebOutputBase
}

// SpoonsWebController はスプーンの Web コントローラー。
type SpoonsWebController = GameWebController[usecase.SpoonsInteractorIF, SpoonsWebInput, *SpoonsWebOutput]

// NewSpoonsWebController and NewSpoonsWebControllerWithProvider are the standard
// and provider-backed constructors for SpoonsWebController.
var NewSpoonsWebController, NewSpoonsWebControllerWithProvider = webControllerPair[usecase.SpoonsInteractorIF, SpoonsWebInput, *SpoonsWebOutput](
	newSpoonsDefaultOutput, spoonsDispatch,
)

func newSpoonsDefaultOutput(msg string) *SpoonsWebOutput {
	return &SpoonsWebOutput{
		WinnerIdx:       -1,
		FirstGrabberIdx: -1,
		RoundLoserIdx:   -1,
		CpuDifficulty:   int(domain.SpoonsCpuNormal),
		Players:         make([]*SpoonsWebPlayer, 0),
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func spoonsDispatch(bc *baseController, w http.ResponseWriter, si usecase.SpoonsInteractorIF, param SpoonsWebInput, _ func(string) *SpoonsWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, si.ResetWithConfig(spoonsConfigFromInput(si.GetConfig(), param.Config)))
		} else {
			bc.writePresenterResponse(w, si.Reset())
		}
		return true
	case "p", "pass":
		bc.writePresenterResponse(w, si.Pass(derefDefault(param.CardIndex, 0)))
		return true
	case "g", "grab":
		bc.writePresenterResponse(w, si.Grab())
		return true
	case "n", "next":
		bc.writePresenterResponse(w, si.NextRound())
		return true
	case "log", "l":
		bc.writePresenterResponse(w, si.ActionLog())
		return true
	}
	return false
}

// spoonsConfigFromInput merges the partial Web config request into the current
// config so missing fields default to existing values rather than zero.
func spoonsConfigFromInput(current domain.SpoonsConfig, in *SpoonsWebConfig) domain.SpoonsConfig {
	out := current
	if in.CpuDifficulty != nil {
		out.CpuDifficulty = domain.SpoonsCpuDifficulty(*in.CpuDifficulty)
	}
	return out
}
