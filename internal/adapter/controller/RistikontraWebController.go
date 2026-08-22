//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// RistikontraWebConfig はローカルルール設定 (入力・出力共用)。
type RistikontraWebConfig struct {
	PlayerCnt     int `json:"playerCnt"`
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts RistikontraWebConfig to domain.RistikontraConfig.
func (c RistikontraWebConfig) ToConfig() domain.RistikontraConfig {
	return domain.RistikontraConfig{
		PlayerCnt:     c.PlayerCnt,
		CpuDifficulty: domain.RistikontraCpuDifficulty(c.CpuDifficulty),
	}
}

// RistikontraWebInput は Pişti Web インプット。
type RistikontraWebInput struct {
	BaseWebInput
	HandIndex int                   `json:"handIndex"`
	Config    *RistikontraWebConfig `json:"config"`
}

// RistikontraWebOutputPlayer は Pişti Web アウトプットプレイヤー。
type RistikontraWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	CardCount     int              `json:"cardCount"`
	Cards         []*WebOutputCard `json:"cards"`
	CapturedCount int              `json:"capturedCount"`
	FinalScore    int              `json:"finalScore"`
}

// RistikontraWebOutput は Pişti Web アウトプット。
type RistikontraWebOutput struct {
	Players        []*RistikontraWebOutputPlayer `json:"players"`
	CurrentTurn    int                           `json:"currentTurn"`
	Pile           []*WebOutputCard              `json:"pile"`
	PileTop        *WebOutputCard                `json:"pileTop"`
	PileCount      int                           `json:"pileCount"`
	LastCaptureIdx int                           `json:"lastCaptureIdx"`
	// CounterRank は打ち返しの対象になっているランク (0 = 対象なし)。
	// このランクを今出せば、直前の捕獲を束ごと奪える。
	CounterRank   int                  `json:"counterRank"`
	GameEndFlag   bool                 `json:"gameEndFlag"`
	Phase         string               `json:"phase"`
	Config        RistikontraWebConfig `json:"config"`
	RemainingDeck int                  `json:"remainingDeck"`
	Winners       []int                `json:"winners"`
	FinalScores   []int                `json:"finalScores"`
	WebOutputBase
}

// RistikontraWebController は Pişti Web コントローラークラス。
type RistikontraWebController = GameWebController[usecase.RistikontraInteractorIF, RistikontraWebInput, *RistikontraWebOutput]

// NewRistikontraWebController, NewRistikontraWebControllerWithProvider are the standard
// and provider-backed constructors for RistikontraWebController.
var NewRistikontraWebController, NewRistikontraWebControllerWithProvider = webControllerPair[usecase.RistikontraInteractorIF, RistikontraWebInput, *RistikontraWebOutput](
	newRistikontraDefaultOutput, ristikontraDispatch,
)

func newRistikontraDefaultOutput(msg string) *RistikontraWebOutput {
	return &RistikontraWebOutput{
		Players:       make([]*RistikontraWebOutputPlayer, 0),
		Pile:          make([]*WebOutputCard, 0),
		Winners:       make([]int, 0),
		FinalScores:   make([]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func ristikontraDispatch(bc *baseController, w http.ResponseWriter, pi usecase.RistikontraInteractorIF, param RistikontraWebInput, _ func(string) *RistikontraWebOutput) bool {
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
