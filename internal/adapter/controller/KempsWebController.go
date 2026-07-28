//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KempsWebInput はケムプスの Web 入力。
type KempsWebInput struct {
	BaseWebInput
	HandIndex  *int            `json:"handIndex,omitempty"`
	FieldIndex *int            `json:"fieldIndex,omitempty"`
	SignalType *int            `json:"signalType,omitempty"`
	TargetSeat *int            `json:"targetSeat,omitempty"`
	Config     *KempsWebConfig `json:"config,omitempty"`
}

// KempsWebConfig はケムプスの設定リクエスト。
type KempsWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// KempsWebPlayer はプレイヤー出力。
type KempsWebPlayer struct {
	Name           string           `json:"name"`
	IsHuman        bool             `json:"isHuman"`
	Team           int              `json:"team"`
	HandSize       int              `json:"handSize"`
	Hand           []*WebOutputCard `json:"hand"`
	HasFourOfAKind bool             `json:"hasFourOfAKind"`
}

// KempsWebOutput はケムプスの Web 出力。
type KempsWebOutput struct {
	Phase             int               `json:"phase"`
	GameEndFlag       bool              `json:"gameEndFlag"`
	WinnerTeam        int               `json:"winnerTeam"`
	CurrentPlayerIdx  int               `json:"currentPlayerIdx"`
	IsHumanTurn       bool              `json:"isHumanTurn"`
	TeamScores        []int             `json:"teamScores"`
	Field             []*WebOutputCard  `json:"field"`
	SignalType        int               `json:"signalType"`
	PartnerSignaling  bool              `json:"partnerSignaling"`
	OpponentSignaling bool              `json:"opponentSignaling"`
	FourHolderIdx     int               `json:"fourHolderIdx"`
	RoundResult       int               `json:"roundResult"`
	RoundWinnerTeam   int               `json:"roundWinnerTeam"`
	RoundNumber       int               `json:"roundNumber"`
	Players           []*KempsWebPlayer `json:"players"`
	CpuDifficulty     int               `json:"cpuDifficulty"`
	TargetScore       int               `json:"targetScore"`
	WebOutputBase
}

// KempsWebController はケムプスの Web コントローラー。
type KempsWebController = GameWebController[usecase.KempsInteractorIF, KempsWebInput, *KempsWebOutput]

// NewKempsWebController and NewKempsWebControllerWithProvider are the standard
// and provider-backed constructors for KempsWebController.
var NewKempsWebController, NewKempsWebControllerWithProvider = webControllerPair[usecase.KempsInteractorIF, KempsWebInput, *KempsWebOutput](
	newKempsDefaultOutput, kempsDispatch,
)

func newKempsDefaultOutput(msg string) *KempsWebOutput {
	return &KempsWebOutput{
		WinnerTeam:      -1,
		FourHolderIdx:   -1,
		RoundWinnerTeam: -1,
		CpuDifficulty:   int(domain.KempsCpuNormal),
		TargetScore:     domain.KempsTargetScore,
		TeamScores:      make([]int, 0),
		Field:           make([]*WebOutputCard, 0),
		Players:         make([]*KempsWebPlayer, 0),
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func kempsDispatch(bc *baseController, w http.ResponseWriter, ki usecase.KempsInteractorIF, param KempsWebInput, _ func(string) *KempsWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, ki.ResetWithConfig(kempsConfigFromInput(ki.GetConfig(), param.Config)))
		} else {
			bc.writePresenterResponse(w, ki.Reset())
		}
		return true
	case "s", "swap":
		bc.writePresenterResponse(w, ki.Swap(derefDefault(param.HandIndex, 0), derefDefault(param.FieldIndex, 0)))
		return true
	case "p", "pass":
		bc.writePresenterResponse(w, ki.Pass())
		return true
	case "sig", "signal":
		bc.writePresenterResponse(w, ki.SetSignal(derefDefault(param.SignalType, 0)))
		return true
	case "k", "kemps":
		bc.writePresenterResponse(w, ki.DeclareKemps())
		return true
	case "c", "counter":
		bc.writePresenterResponse(w, ki.DeclareCounterKemps(derefDefault(param.TargetSeat, 0)))
		return true
	case "n", "next":
		bc.writePresenterResponse(w, ki.NextRound())
		return true
	case "log", "l":
		bc.writePresenterResponse(w, ki.ActionLog())
		return true
	}
	return false
}

// kempsConfigFromInput merges the partial Web config request into the current
// config so missing fields default to existing values rather than zero.
func kempsConfigFromInput(current domain.KempsConfig, in *KempsWebConfig) domain.KempsConfig {
	out := current
	if in.CpuDifficulty != nil {
		out.CpuDifficulty = domain.KempsCpuDifficulty(*in.CpuDifficulty)
	}
	if in.TargetScore != nil {
		out.TargetScore = *in.TargetScore
	}
	return out
}
