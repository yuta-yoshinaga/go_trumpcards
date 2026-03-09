package controller

import (
	"fmt"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// DoubtWebInput ダウトWebインプット
type DoubtWebInput struct {
	BaseWebInput
	CardIndices          []int `json:"cardIndices,omitempty"`
	ClaimedValue         int   `json:"claimedValue,omitempty"`
	DoubterIndices       []int `json:"doubterIndices,omitempty"`
	DoubtWindowSec       *int  `json:"doubtWindowSec,omitempty"`
	CpuMemoryLevel       *int  `json:"cpuMemoryLevel,omitempty"`
	PenaltyDrawLimit     *int  `json:"penaltyDrawLimit,omitempty"`
	CpuHesitationEnabled bool  `json:"cpuHesitationEnabled,omitempty"`
	CpuMetaAI            bool  `json:"cpuMetaAI,omitempty"`
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
	HesitationMs int  `json:"hesitationMs,omitempty"`
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
	MetaAI           *DoubtWebOutputMetaAI      `json:"metaAI,omitempty"`
}

// DoubtWebOutputMetaAI メタAI情報
type DoubtWebOutputMetaAI struct {
	Enabled       bool    `json:"enabled"`
	GamesPlayed   int     `json:"gamesPlayed"`
	BluffRate     float64 `json:"bluffRate"`
	DoubtAccuracy float64 `json:"doubtAccuracy"`
}

// DoubtWebController ダウトWebコントローラークラス
type DoubtWebController = GameWebController[usecase.DoubtInteractorIF, DoubtWebInput, *DoubtWebOutput]

// MaxCardIndices カードインデックスの最大数 (52枚デッキ)
const MaxCardIndices = 52

// NewDoubtWebController コンストラクタ
func NewDoubtWebController(factory func() usecase.DoubtInteractorIF) *DoubtWebController {
	return NewGameWebController(factory, newDoubtDefaultOutput, doubtDispatch)
}

func newDoubtDefaultOutput(msg string) *DoubtWebOutput {
	return &DoubtWebOutput{
		Players:     make([]*DoubtWebOutputPlayer, 0),
		CpuDoubters: make([]int, 0),
		CpuActions:  make([]*DoubtWebOutputAction, 0),
		WinnerIdx:   -1,
		Message:     msg,
	}
}

func doubtDispatch(bc *baseController, w rest.ResponseWriter, dgi usecase.DoubtInteractorIF, param DoubtWebInput, newDefault func(string) *DoubtWebOutput) bool {
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
		cfg.CpuHesitationEnabled = param.CpuHesitationEnabled
		cfg.CpuMetaAI = param.CpuMetaAI
		bc.writePresenterResponse(w, dgi.ResetWithConfig(cfg))
	case "p", "play":
		if len(param.CardIndices) > MaxCardIndices {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error."))
			return true
		}
		if param.ClaimedValue < domain.MinClaimedValue || param.ClaimedValue > domain.MaxClaimedValue {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(fmt.Sprintf("param error: claimedValue must be between %d and %d.", domain.MinClaimedValue, domain.MaxClaimedValue)))
			return true
		}
		bc.writePresenterResponse(w, dgi.Play(param.CardIndices, param.ClaimedValue))
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
		bc.writePresenterResponse(w, dgi.ResolveDoubt(doubters))
	case "rp", "reset-profile":
		bc.writePresenterResponse(w, dgi.ResetProfile())
	case "s", "skip":
		cpuDoubters := dgi.GetCpuDoubters()
		if len(cpuDoubters) > 0 {
			bc.writePresenterResponse(w, dgi.ResolveDoubt(cpuDoubters))
		} else {
			bc.writePresenterResponse(w, dgi.SkipDoubt())
		}
	case "log", "l":
		bc.writePresenterResponse(w, dgi.ActionLog())
	default:
		return false
	}
	return true
}
