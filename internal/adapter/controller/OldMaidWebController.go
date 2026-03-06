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
}

// OldMaidWebOutput ババ抜きWebアウトプット
type OldMaidWebOutput struct {
	Players               []*OldMaidWebOutputPlayer    `json:"players"`
	CurrentTurn           int                          `json:"currentTurn"`
	NextDrawTargetIdx     int                          `json:"nextDrawTargetIdx"`
	GameEndFlag           bool                         `json:"gameEndFlag"`
	LoserIdx              int                          `json:"loserIdx"`
	LastDrawPlayerIdx     int                          `json:"lastDrawPlayerIdx"`
	LastDrawFromIdx       int                          `json:"lastDrawFromIdx"`
	LastDrawCard          *WebOutputCard               `json:"lastDrawCard"`
	LastDiscardedPairs    int                          `json:"lastDiscardedPairs"`
	LastDiscardedCards    []*WebOutputCard             `json:"lastDiscardedCards"`
	HasDrawn              bool                         `json:"hasDrawn"`
	CpuActions            []*OldMaidWebOutputCpuAction `json:"cpuActions"`
	HumanAction           *OldMaidWebOutputCpuAction   `json:"humanAction"`
	CpuHighlightedCardIdx int                          `json:"cpuHighlightedCardIdx"`
	RemovedCard           *WebOutputCard               `json:"removedCard"`
	Mode                  int                          `json:"mode"`
	Message               string                       `json:"message"`
	MessageCode           string                       `json:"messageCode,omitempty"`
	MessageParams         map[string]string            `json:"messageParams,omitempty"`
}

// OldMaidWebController ババ抜きWebコントローラークラス
type OldMaidWebController struct {
	baseController
	factory func() usecase.OldMaidInteractorIF
	store   *SessionStore[usecase.OldMaidInteractorIF]
}

// NewOldMaidWebController コンストラクタ
func NewOldMaidWebController(factory func() usecase.OldMaidInteractorIF) *OldMaidWebController {
	return &OldMaidWebController{
		factory: factory,
		store:   NewSessionStore[usecase.OldMaidInteractorIF](),
	}
}

// Exec ゲーム実行
func (owc *OldMaidWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	execWithSession(&owc.baseController, w, r, owc.store, owc.factory,
		func(msg string) any { return owc.newDefaultOutput(msg) },
		nil,
		func(w rest.ResponseWriter, omi usecase.OldMaidInteractorIF, param OldMaidWebInput) bool {
			switch param.Command {
			case "r", "reset":
				if param.Mode < 0 || param.Mode > int(domain.OldMaidModeJijiNuki) {
					owc.writeJsonResponse(w, http.StatusBadRequest, owc.newDefaultOutput("param error: mode must be between 0 and 1."))
					return true
				}
				cfg := domain.OldMaidConfig{
					Mode:                 domain.OldMaidMode(param.Mode),
					CpuPlacementStrategy: param.CpuPlacementStrategy,
					CpuMemoryAI:          param.CpuMemoryAI,
				}
				owc.writePresenterResponse(w, omi.Reset(cfg))
			case "d", "draw":
				drawIdx := -1
				if param.DrawIdx != nil {
					drawIdx = *param.DrawIdx
				}
				owc.writePresenterResponse(w, omi.Draw(drawIdx))
			case "s", "shuffle":
				owc.writePresenterResponse(w, omi.Shuffle())
			case "reorder":
				if param.ReorderIndices == nil {
					owc.writeJsonResponse(w, http.StatusBadRequest, owc.newDefaultOutput("param error: reorderIndices is required."))
					return true
				}
				owc.writePresenterResponse(w, omi.Reorder(param.ReorderIndices))
			default:
				return false
			}
			return true
		})
}

// Stop stops the background cleanup goroutine of the session store.
func (owc *OldMaidWebController) Stop() {
	owc.store.Stop()
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (owc *OldMaidWebController) newDefaultOutput(msg string) *OldMaidWebOutput {
	return &OldMaidWebOutput{
		Players:               make([]*OldMaidWebOutputPlayer, 0),
		CpuActions:            make([]*OldMaidWebOutputCpuAction, 0),
		CpuHighlightedCardIdx: -1,
		Message:               msg,
	}
}
