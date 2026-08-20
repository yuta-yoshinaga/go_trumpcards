package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CassinoWebConfig ローカルルール設定 (入力・出力共用)
type CassinoWebConfig struct {
	TargetScore       int  `json:"targetScore"`
	MultiBuildEnabled bool `json:"multiBuildEnabled"`
	SweepBonusEnabled bool `json:"sweepBonusEnabled"`
	CpuDifficulty     int  `json:"cpuDifficulty"`
}

// ToConfig converts CassinoWebConfig to domain.CassinoConfig.
func (c CassinoWebConfig) ToConfig() domain.CassinoConfig {
	return domain.CassinoConfig{
		TargetScore:       c.TargetScore,
		MultiBuildEnabled: c.MultiBuildEnabled,
		SweepBonusEnabled: c.SweepBonusEnabled,
		CpuDifficulty:     domain.CassinoCpuDifficulty(c.CpuDifficulty),
	}
}

// CassinoWebInput カシノWebインプット
type CassinoWebInput struct {
	BaseWebInput
	HandIndex     int               `json:"handIndex"`
	TableIndices  []int             `json:"tableIndices"`
	BuildIndices  []int             `json:"buildIndices"`
	DeclaredValue int               `json:"declaredValue"`
	Config        *CassinoWebConfig `json:"config"`
}

// CassinoWebOutputPlayer カシノWebアウトプットプレイヤー
type CassinoWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	CardCount     int              `json:"cardCount"`
	Cards         []*WebOutputCard `json:"cards"`
	CapturedCount int              `json:"capturedCount"`
	SweepCount    int              `json:"sweepCount"`
	TotalScore    int              `json:"totalScore"`
}

// CassinoWebOutputBuild 場のビルド
type CassinoWebOutputBuild struct {
	OwnerIdx int                `json:"ownerIdx"`
	Value    int                `json:"value"`
	Groups   [][]*WebOutputCard `json:"groups"`
	IsMulti  bool               `json:"isMulti"`
}

// CassinoWebOutputAction 行動記録
type CassinoWebOutputAction struct {
	PlayerIdx     int              `json:"playerIdx"`
	Type          string           `json:"type"`
	PlayedCard    *WebOutputCard   `json:"playedCard"`
	CapturedCards []*WebOutputCard `json:"capturedCards"`
	BuildValue    int              `json:"buildValue"`
	IsSweep       bool             `json:"isSweep"`
}

// CassinoWebOutputScoreDetail 得点内訳
type CassinoWebOutputScoreDetail struct {
	Cards           map[int]int `json:"cards"`
	Spades          map[int]int `json:"spades"`
	Aces            map[int]int `json:"aces"`
	HasBigCasino    int         `json:"hasBigCasino"`
	HasLittleCasino int         `json:"hasLittleCasino"`
	Sweeps          map[int]int `json:"sweeps"`
	Gained          map[int]int `json:"gained"`
}

// CassinoWebOutput カシノWebアウトプット
type CassinoWebOutput struct {
	Players         []*CassinoWebOutputPlayer    `json:"players"`
	CurrentTurn     int                          `json:"currentTurn"`
	TableCards      []*WebOutputCard             `json:"tableCards"`
	Builds          []*CassinoWebOutputBuild     `json:"builds"`
	LastCaptureIdx  int                          `json:"lastCaptureIdx"`
	GameEndFlag     bool                         `json:"gameEndFlag"`
	Phase           string                       `json:"phase"`
	Config          CassinoWebConfig             `json:"config"`
	CpuActions      []*CassinoWebOutputAction    `json:"cpuActions"`
	HumanAction     *CassinoWebOutputAction      `json:"humanAction"`
	RemainingDeck   int                          `json:"remainingDeck"`
	PacksDealt      int                          `json:"packsDealt"`
	RoundWinners    []int                        `json:"roundWinners"`
	LastRoundDetail *CassinoWebOutputScoreDetail `json:"lastRoundDetail"`
	WebOutputBase
}

// CassinoWebController カシノWebコントローラークラス
type CassinoWebController = GameWebController[usecase.CassinoInteractorIF, CassinoWebInput, *CassinoWebOutput]

// NewCassinoWebController, NewCassinoWebControllerWithProvider are the
// standard and provider-backed constructors for CassinoWebController.
var NewCassinoWebController, NewCassinoWebControllerWithProvider = webControllerPair[usecase.CassinoInteractorIF, CassinoWebInput, *CassinoWebOutput](
	newCassinoDefaultOutput, cassinoDispatch,
)

func newCassinoDefaultOutput(msg string) *CassinoWebOutput {
	return &CassinoWebOutput{
		Players:       make([]*CassinoWebOutputPlayer, 0),
		TableCards:    make([]*WebOutputCard, 0),
		Builds:        make([]*CassinoWebOutputBuild, 0),
		CpuActions:    make([]*CassinoWebOutputAction, 0),
		RoundWinners:  make([]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

// cassinoMaxIndices caps client-supplied index slices for take / build commands.
// 52 is the absolute physical maximum (single deck), and using 64 provides a
// modest safety margin while still being well within reasonable request size.
const cassinoMaxIndices = 64

func cassinoDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CassinoInteractorIF, param CassinoWebInput, defaultOutput func(string) *CassinoWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, ci.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, ci.Reset())
		}
	case "n", "next":
		bc.writePresenterResponse(w, ci.NextRound())
	case "t", "take":
		if len(param.TableIndices) > cassinoMaxIndices || len(param.BuildIndices) > cassinoMaxIndices {
			bc.writeJsonResponse(w, http.StatusBadRequest, defaultOutput("too many indices"))
			return true
		}
		bc.writePresenterResponse(w, ci.Take(param.HandIndex, param.TableIndices, param.BuildIndices))
	case "b", "build":
		if len(param.TableIndices) > cassinoMaxIndices {
			bc.writeJsonResponse(w, http.StatusBadRequest, defaultOutput("too many table indices"))
			return true
		}
		bc.writePresenterResponse(w, ci.Build(param.HandIndex, param.TableIndices, param.DeclaredValue))
	case "tr", "trail":
		bc.writePresenterResponse(w, ci.Trail(param.HandIndex))
	default:
		return dispatchHintAndLog(param.Command, bc, w, ci.Hint, ci.ActionLog)
	}
	return true
}
