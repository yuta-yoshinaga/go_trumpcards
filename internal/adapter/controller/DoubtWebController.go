package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// DoubtWebInput ダウトWebインプット
type DoubtWebInput struct {
	BaseWebInput
	CardIndices      []int `json:"cardIndices,omitempty"`
	ClaimedValue     int   `json:"claimedValue,omitempty"`
	DoubterIndices   []int `json:"doubterIndices,omitempty"`
	DoubtWindowSec   *int  `json:"doubtWindowSec,omitempty"`
	CpuMemoryLevel   *int  `json:"cpuMemoryLevel,omitempty"`
	PenaltyDrawLimit *int  `json:"penaltyDrawLimit,omitempty"`
}

// DoubtWebOutputPlayer ダウトWebアウトプットプレイヤー
type DoubtWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// DoubtWebOutputAction ダウトのプレイヤー行動記録
type DoubtWebOutputAction struct {
	PlayerIdx    int  `json:"playerIdx"`
	ClaimedValue int  `json:"claimedValue"`
	CardCount    int  `json:"cardCount"`
	IsBluff      bool `json:"isBluff"`
	HasTell      bool `json:"hasTell"`
}

// DoubtWebOutputDoubtResult ダウト解決結果
type DoubtWebOutputDoubtResult struct {
	DoubterIdx     int              `json:"doubterIdx"`
	CardPlayerIdx  int              `json:"cardPlayerIdx"`
	WasLying       bool             `json:"wasLying"`
	LoserIdx       int              `json:"loserIdx"`
	CardCount      int              `json:"cardCount"`
	DiscardedCount int              `json:"discardedCount"`
	RevealedCards  []*WebOutputCard `json:"revealedCards"`
}

// DoubtWebOutput ダウトWebアウトプット
type DoubtWebOutput struct {
	Players          []*DoubtWebOutputPlayer    `json:"players"`
	CurrentTurn      int                        `json:"currentTurn"`
	Phase            int                        `json:"phase"`
	TableCardCount   int                        `json:"tableCardCount"`
	LastAction       *DoubtWebOutputAction      `json:"lastAction"`
	CpuDoubters      []int                      `json:"cpuDoubters"`
	CpuActions       []*DoubtWebOutputAction    `json:"cpuActions"`
	HumanAction      *DoubtWebOutputAction      `json:"humanAction"`
	LastDoubtResult  *DoubtWebOutputDoubtResult `json:"lastDoubtResult"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	Message          string                     `json:"message"`
	MessageCode      string                     `json:"messageCode,omitempty"`
	MessageParams    map[string]string          `json:"messageParams,omitempty"`
	DoubtWindowSec   int                        `json:"doubtWindowSec"`
	PenaltyDrawLimit int                        `json:"penaltyDrawLimit"`
}

// DoubtWebController ダウトWebコントローラークラス
type DoubtWebController struct {
	baseController
	factory func() usecase.DoubtInteractorIF
	store   *SessionStore[usecase.DoubtInteractorIF]
}

// NewDoubtWebController コンストラクタ
func NewDoubtWebController(factory func() usecase.DoubtInteractorIF) *DoubtWebController {
	return &DoubtWebController{
		factory: factory,
		store:   NewSessionStore[usecase.DoubtInteractorIF](),
	}
}

// MaxCardIndices カードインデックスの最大数 (52枚デッキ)
const MaxCardIndices = 52

// errCardIndicesOverflow is returned by the Doubt validate callback when CardIndices exceeds the limit.
var errCardIndicesOverflow = errors.New("card indices overflow")

// Exec ゲーム実行
func (dwc *DoubtWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	execWithSession(&dwc.baseController, w, r, dwc.store, dwc.factory,
		func(msg string) any { return dwc.newDefaultOutput(msg) },
		func(param DoubtWebInput) error {
			if len(param.CardIndices) > MaxCardIndices {
				return errCardIndicesOverflow
			}
			return nil
		},
		func(w rest.ResponseWriter, dgi usecase.DoubtInteractorIF, param DoubtWebInput) bool {
			switch param.Command {
			case "r", "reset":
				cfg := domain.DefaultDoubtConfig()
				if param.DoubtWindowSec != nil && *param.DoubtWindowSec >= 1 {
					cfg.DoubtWindowSec = *param.DoubtWindowSec
				}
				if param.CpuMemoryLevel != nil {
					level := *param.CpuMemoryLevel
					if level >= int(domain.DoubtMemoryLevelEasy) && level <= int(domain.DoubtMemoryLevelHard) {
						cfg.CpuMemoryLevel = domain.DoubtMemoryLevel(level)
					}
				}
				if param.PenaltyDrawLimit != nil && *param.PenaltyDrawLimit >= 0 {
					cfg.PenaltyDrawLimit = *param.PenaltyDrawLimit
				}
				dwc.writePresenterResponse(w, dgi.ResetWithConfig(cfg))
			case "p", "play":
				if param.ClaimedValue < domain.MinClaimedValue || param.ClaimedValue > domain.MaxClaimedValue {
					dwc.writeJsonResponse(w, http.StatusBadRequest, dwc.newDefaultOutput(fmt.Sprintf("param error: claimedValue must be between %d and %d.", domain.MinClaimedValue, domain.MaxClaimedValue)))
					return true
				}
				dwc.writePresenterResponse(w, dgi.Play(param.CardIndices, param.ClaimedValue))
			case "d", "doubt":
				cpuDoubters := dgi.GetCpuDoubters()
				humanDoubts := false
				for _, idx := range param.DoubterIndices {
					if idx == 0 {
						humanDoubts = true
						break
					}
				}
				var doubters []int
				if humanDoubts {
					doubters = append([]int{0}, cpuDoubters...)
				} else {
					doubters = cpuDoubters
				}
				dwc.writePresenterResponse(w, dgi.ResolveDoubt(doubters))
			case "s", "skip":
				cpuDoubters := dgi.GetCpuDoubters()
				if len(cpuDoubters) > 0 {
					dwc.writePresenterResponse(w, dgi.ResolveDoubt(cpuDoubters))
				} else {
					dwc.writePresenterResponse(w, dgi.SkipDoubt())
				}
			default:
				return false
			}
			return true
		})
}

// Stop stops the background cleanup goroutine of the session store.
func (dwc *DoubtWebController) Stop() {
	dwc.store.Stop()
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (dwc *DoubtWebController) newDefaultOutput(msg string) *DoubtWebOutput {
	return &DoubtWebOutput{
		Players:     make([]*DoubtWebOutputPlayer, 0),
		CpuDoubters: make([]int, 0),
		CpuActions:  make([]*DoubtWebOutputAction, 0),
		WinnerIdx:   -1,
		Message:     msg,
	}
}
