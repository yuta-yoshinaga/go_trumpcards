//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ScoponeWebConfig ローカルルール設定 (入力・出力共用)
type ScoponeWebConfig struct {
	TargetScore   int `json:"targetScore"`
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig converts ScoponeWebConfig to domain.ScoponeConfig.
func (c ScoponeWebConfig) ToConfig() domain.ScoponeConfig {
	return domain.ScoponeConfig{
		TargetScore:   c.TargetScore,
		CpuDifficulty: domain.ScoponeCpuDifficulty(c.CpuDifficulty),
	}
}

// ScoponeWebInput スコポーネWebインプット
type ScoponeWebInput struct {
	BaseWebInput
	HandIndex    int               `json:"handIndex"`
	TableIndices []int             `json:"tableIndices"`
	Config       *ScoponeWebConfig `json:"config"`
}

// ScoponeWebOutputPlayer スコポーネWebアウトプットプレイヤー
type ScoponeWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	Team          int              `json:"team"`
	HandCount     int              `json:"handCount"`
	Cards         []*WebOutputCard `json:"cards"`
	CapturedCount int              `json:"capturedCount"`
	ScopaCount    int              `json:"scopaCount"`
}

// ScoponeWebOutputScoreDetail 得点内訳 (チーム単位)
type ScoponeWebOutputScoreDetail struct {
	Cards      [domain.ScoponeTeamCnt]int `json:"cards"`
	Diamonds   [domain.ScoponeTeamCnt]int `json:"diamonds"`
	Sevens     [domain.ScoponeTeamCnt]int `json:"sevens"`
	Scopas     [domain.ScoponeTeamCnt]int `json:"scopas"`
	Gained     [domain.ScoponeTeamCnt]int `json:"gained"`
	Settebello int                        `json:"settebello"`
}

// ScoponeWebOutput スコポーネWebアウトプット
type ScoponeWebOutput struct {
	Players         []*ScoponeWebOutputPlayer    `json:"players"`
	TableCards      []*WebOutputCard             `json:"tableCards"`
	Phase           string                       `json:"phase"`
	RoundNumber     int                          `json:"roundNumber"`
	CurrentTurn     int                          `json:"currentTurn"`
	DealerIdx       int                          `json:"dealerIdx"`
	TeamScores      []int                        `json:"teamScores"`
	LastCaptureIdx  int                          `json:"lastCaptureIdx"`
	WinnerTeam      int                          `json:"winnerTeam"`
	GameEndFlag     bool                         `json:"gameEndFlag"`
	IsHumanTurn     bool                         `json:"isHumanTurn"`
	HandCaptures    [][][]int                    `json:"handCaptures"`
	LastRoundDetail *ScoponeWebOutputScoreDetail `json:"lastRoundDetail"`
	Config          ScoponeWebConfig             `json:"config"`
	WebOutputBase
}

// ScoponeWebController スコポーネWebコントローラークラス
type ScoponeWebController = GameWebController[usecase.ScoponeInteractorIF, ScoponeWebInput, *ScoponeWebOutput]

// NewScoponeWebController, NewScoponeWebControllerWithProvider are the standard and
// provider-backed constructors for ScoponeWebController.
var NewScoponeWebController, NewScoponeWebControllerWithProvider = webControllerPair[usecase.ScoponeInteractorIF, ScoponeWebInput, *ScoponeWebOutput](
	newScoponeDefaultOutput, scoponeDispatch,
)

func newScoponeDefaultOutput(msg string) *ScoponeWebOutput {
	return &ScoponeWebOutput{
		Players:       make([]*ScoponeWebOutputPlayer, 0),
		TableCards:    make([]*WebOutputCard, 0),
		TeamScores:    make([]int, 0),
		HandCaptures:  make([][][]int, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

// scoponeMaxIndices caps client-supplied table-index slices for the play command.
const scoponeMaxIndices = 40

func scoponeDispatch(bc *baseController, w http.ResponseWriter, si usecase.ScoponeInteractorIF, param ScoponeWebInput, defaultOutput func(string) *ScoponeWebOutput) bool {
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
		if len(param.TableIndices) > scoponeMaxIndices {
			bc.writeJsonResponse(w, http.StatusBadRequest, defaultOutput("too many indices"))
			return true
		}
		bc.writePresenterResponse(w, si.Play(param.HandIndex, param.TableIndices))
	case "config":
		if param.Config != nil {
			bc.writePresenterResponse(w, si.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, si.Reset())
		}
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
