package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// OldMaidWebInput ババ抜きWebインプット
type OldMaidWebInput struct {
	BaseWebInput
	DrawIdx              *int  `json:"drawIdx"` // 引くカードのインデックス。nil の場合はランダム選択。
	ReorderIndices       []int `json:"reorderIndices"`
	Mode                 int   `json:"mode"`
	CpuPlacementStrategy bool  `json:"cpuPlacementStrategy"`
	CpuMemoryAI          bool  `json:"cpuMemoryAI"`
	CpuHesitationEnabled bool  `json:"cpuHesitationEnabled"`
}

// OldMaidWebOutputPlayer ババ抜きWebアウトプットプレイヤー
type OldMaidWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// OldMaidWebOutputCpuAction CPUターンの行動記録
type OldMaidWebOutputCpuAction struct {
	DrawPlayerIdx  int              `json:"drawPlayerIdx"`
	DrawFromIdx    int              `json:"drawFromIdx"`
	DrawnCard      *WebOutputCard   `json:"drawnCard"`
	DiscardedPairs int              `json:"discardedPairs"`
	DiscardedCards []*WebOutputCard `json:"discardedCards"`
	HesitationMs   int              `json:"hesitationMs"`
}

// OldMaidWebOutputDrawHistoryEntry ゲーム全体の引き履歴エントリ
type OldMaidWebOutputDrawHistoryEntry struct {
	DrawPlayerIdx  int  `json:"drawPlayerIdx"`
	DrawFromIdx    int  `json:"drawFromIdx"`
	DiscardedPairs int  `json:"discardedPairs"`
	DrawerFinished bool `json:"drawerFinished"`
	TargetFinished bool `json:"targetFinished"`
}

// OldMaidWebOutput ババ抜きWebアウトプット
type OldMaidWebOutput struct {
	Players               []*OldMaidWebOutputPlayer           `json:"players"`
	CurrentTurn           int                                 `json:"currentTurn"`
	NextDrawTargetIdx     int                                 `json:"nextDrawTargetIdx"`
	GameEndFlag           bool                                `json:"gameEndFlag"`
	LoserIdx              int                                 `json:"loserIdx"`
	LastDrawPlayerIdx     int                                 `json:"lastDrawPlayerIdx"`
	LastDrawFromIdx       int                                 `json:"lastDrawFromIdx"`
	LastDrawCard          *WebOutputCard                      `json:"lastDrawCard"`
	LastDiscardedPairs    int                                 `json:"lastDiscardedPairs"`
	LastDiscardedCards    []*WebOutputCard                    `json:"lastDiscardedCards"`
	HasDrawn              bool                                `json:"hasDrawn"`
	CpuActions            []*OldMaidWebOutputCpuAction        `json:"cpuActions"`
	HumanAction           *OldMaidWebOutputCpuAction          `json:"humanAction"`
	DrawHistory           []*OldMaidWebOutputDrawHistoryEntry `json:"drawHistory"`
	CpuHighlightedCardIdx int                                 `json:"cpuHighlightedCardIdx"`
	RemovedCard           *WebOutputCard                      `json:"removedCard"`
	Mode                  int                                 `json:"mode"`
	Message               string                              `json:"message"`
	MessageCode           string                              `json:"messageCode,omitempty"`
	MessageParams         map[string]string                   `json:"messageParams,omitempty"`
}

// OldMaidWebController ババ抜きWebコントローラークラス
type OldMaidWebController = GameWebController[usecase.OldMaidInteractorIF, OldMaidWebInput, *OldMaidWebOutput]

// NewOldMaidWebController コンストラクタ
func NewOldMaidWebController(factory func() usecase.OldMaidInteractorIF) *OldMaidWebController {
	return NewGameWebController(factory, newOldMaidDefaultOutput, oldMaidDispatch)
}

func newOldMaidDefaultOutput(msg string) *OldMaidWebOutput {
	return &OldMaidWebOutput{
		Players:               make([]*OldMaidWebOutputPlayer, 0),
		CpuActions:            make([]*OldMaidWebOutputCpuAction, 0),
		DrawHistory:           make([]*OldMaidWebOutputDrawHistoryEntry, 0),
		CpuHighlightedCardIdx: -1,
		Message:               msg,
	}
}

func oldMaidDispatch(bc *baseController, w rest.ResponseWriter, omi usecase.OldMaidInteractorIF, param OldMaidWebInput, newDefault func(string) *OldMaidWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Mode < 0 || param.Mode > int(domain.OldMaidModeJijiNuki) {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: mode must be between 0 and 1."))
			return true
		}
		cfg := domain.OldMaidConfig{
			Mode:                 domain.OldMaidMode(param.Mode),
			CpuPlacementStrategy: param.CpuPlacementStrategy,
			CpuMemoryAI:          param.CpuMemoryAI,
			CpuHesitationEnabled: param.CpuHesitationEnabled,
		}
		bc.writePresenterResponse(w, omi.Reset(cfg))
	case "d", "draw":
		drawIdx := derefIntDefault(param.DrawIdx, -1)
		bc.writePresenterResponse(w, omi.Draw(drawIdx))
	case "s", "shuffle":
		bc.writePresenterResponse(w, omi.Shuffle())
	case "reorder":
		if param.ReorderIndices == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: reorderIndices is required."))
			return true
		}
		bc.writePresenterResponse(w, omi.Reorder(param.ReorderIndices))
	default:
		return false
	}
	return true
}
