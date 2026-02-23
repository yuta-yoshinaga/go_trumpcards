package controller

import (
	"log"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// OldMaidWebInput ババ抜きWebインプット
type OldMaidWebInput struct {
	Command   string `json:"command"`
	DrawIdx   *int   `json:"drawIdx"` // 引くカードのインデックス。nil の場合はランダム選択。
	SessionId string `json:"sessionId"`
}

// OldMaidWebOutputCard ババ抜きWebアウトプットカード
type OldMaidWebOutputCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}

// OldMaidWebOutputPlayer ババ抜きWebアウトプットプレイヤー
type OldMaidWebOutputPlayer struct {
	ID         int                     `json:"id"`
	IsHuman    bool                    `json:"isHuman"`
	IsFinished bool                    `json:"isFinished"`
	CardCount  int                     `json:"cardCount"`
	Cards      []*OldMaidWebOutputCard `json:"cards"`
}

// OldMaidWebOutputCpuAction CPUターンの行動記録
type OldMaidWebOutputCpuAction struct {
	DrawPlayerIdx  int                     `json:"drawPlayerIdx"`
	DrawFromIdx    int                     `json:"drawFromIdx"`
	DrawnCard      *OldMaidWebOutputCard   `json:"drawnCard"`
	DiscardedPairs int                     `json:"discardedPairs"`
	DiscardedCards []*OldMaidWebOutputCard `json:"discardedCards"`
}

// OldMaidWebOutput ババ抜きWebアウトプット
type OldMaidWebOutput struct {
	Players            []*OldMaidWebOutputPlayer    `json:"players"`
	CurrentTurn        int                          `json:"currentTurn"`
	NextDrawTargetIdx  int                          `json:"nextDrawTargetIdx"`
	GameEndFlag        bool                         `json:"gameEndFlag"`
	LoserIdx           int                          `json:"loserIdx"`
	LastDrawPlayerIdx  int                          `json:"lastDrawPlayerIdx"`
	LastDrawFromIdx    int                          `json:"lastDrawFromIdx"`
	LastDrawCard       *OldMaidWebOutputCard        `json:"lastDrawCard"`
	LastDiscardedPairs int                          `json:"lastDiscardedPairs"`
	LastDiscardedCards []*OldMaidWebOutputCard      `json:"lastDiscardedCards"`
	HasDrawn           bool                         `json:"hasDrawn"`
	CpuActions         []*OldMaidWebOutputCpuAction `json:"cpuActions"`
	HumanAction        *OldMaidWebOutputCpuAction   `json:"humanAction"`
	Message            string                       `json:"message"`
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
	var param OldMaidWebInput
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(owc.newDefaultOutput("param error.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	if param.Command == "q" || param.Command == "quit" {
		w.WriteHeader(http.StatusOK)
		if err := w.WriteJson(owc.newDefaultOutput("bye.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	omi, mu, ok := owc.store.GetWithLock(param.SessionId, owc.factory)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(owc.newDefaultOutput("param error.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	drawIdx := -1
	if param.DrawIdx != nil {
		drawIdx = *param.DrawIdx
	}
	mu.Lock()
	defer mu.Unlock()
	errOutput := owc.newDefaultOutput("error.")
	switch param.Command {
	case "r", "reset":
		owc.writePresenterResponse(w, omi.Reset(), errOutput)
	case "d", "draw":
		owc.writePresenterResponse(w, omi.Draw(drawIdx), errOutput)
	default:
		w.WriteHeader(http.StatusOK)
		if err := w.WriteJson(owc.newDefaultOutput("Unsupported command.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
	}
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (owc *OldMaidWebController) newDefaultOutput(msg string) *OldMaidWebOutput {
	return &OldMaidWebOutput{
		Players:    make([]*OldMaidWebOutputPlayer, 0),
		CpuActions: make([]*OldMaidWebOutputCpuAction, 0),
		Message:    msg,
	}
}
