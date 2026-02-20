package controllers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

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
	ID         int                    `json:"id"`
	IsHuman    bool                   `json:"isHuman"`
	IsFinished bool                   `json:"isFinished"`
	CardCount  int                    `json:"cardCount"`
	Cards      []*OldMaidWebOutputCard `json:"cards"`
}

// OldMaidWebOutputCpuAction CPUターンの行動記録
type OldMaidWebOutputCpuAction struct {
	DrawPlayerIdx  int                   `json:"drawPlayerIdx"`
	DrawFromIdx    int                   `json:"drawFromIdx"`
	DrawnCard      *OldMaidWebOutputCard `json:"drawnCard"`
	DiscardedPairs int                   `json:"discardedPairs"`
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
	HasDrawn           bool                         `json:"hasDrawn"`
	CpuActions         []*OldMaidWebOutputCpuAction `json:"cpuActions"`
	Message            string                       `json:"message"`
}

// OldMaidWebController ババ抜きWebコントローラークラス
type OldMaidWebController struct {
	factory  func() usecases.OldMaidInteractorIF
	sessions map[string]usecases.OldMaidInteractorIF
	mu       sync.Mutex
}

// NewOldMaidWebController コンストラクタ
func NewOldMaidWebController(factory func() usecases.OldMaidInteractorIF) *OldMaidWebController {
	return &OldMaidWebController{
		factory:  factory,
		sessions: make(map[string]usecases.OldMaidInteractorIF),
	}
}

// getOrCreateSession セッションIDに対応するインタラクターを取得または生成する
func (owc *OldMaidWebController) getOrCreateSession(sessionId string) usecases.OldMaidInteractorIF {
	owc.mu.Lock()
	defer owc.mu.Unlock()
	omi, ok := owc.sessions[sessionId]
	if !ok {
		omi = owc.factory()
		owc.sessions[sessionId] = omi
	}
	return omi
}

// Exec ゲーム実行
func (owc *OldMaidWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param OldMaidWebInput
	status := http.StatusOK
	responseStr := ""
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" {
		status = http.StatusBadRequest
		responseStr = `{"message":"param error."}`
	} else {
		drawIdx := -1
		if param.DrawIdx != nil {
			drawIdx = *param.DrawIdx
		}
		switch param.Command {
		case "q", "quit":
			responseStr = `{"message":"bye."}`
		case "r", "reset":
			responseStr = owc.getOrCreateSession(param.SessionId).Reset()
		case "d", "draw":
			responseStr = owc.getOrCreateSession(param.SessionId).Draw(drawIdx)
		default:
			responseStr = `{"message":"Unsupported command."}`
		}
	}
	response := new(OldMaidWebOutput)
	response.Players = make([]*OldMaidWebOutputPlayer, 0)
	response.CpuActions = make([]*OldMaidWebOutputCpuAction, 0)
	err = json.Unmarshal([]byte(responseStr), &response)
	if err != nil || responseStr == "" {
		status = http.StatusBadRequest
		response.Message = "error."
	}
	w.WriteHeader(status)
	_ = w.WriteJson(response)
}
