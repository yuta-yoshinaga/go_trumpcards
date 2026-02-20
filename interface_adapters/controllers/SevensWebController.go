package controllers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// SevensWebInput 7並べWebインプット
type SevensWebInput struct {
	Command   string `json:"command"`
	Index     int    `json:"index"` // 出すカードのインデックス。play コマンド用。-1 でパス。
	SessionId string `json:"sessionId"`
}

// SevensWebOutputCard 7並べWebアウトプットカード
type SevensWebOutputCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}

// SevensWebOutputPlayer 7並べWebアウトプットプレイヤー
type SevensWebOutputPlayer struct {
	ID         int                    `json:"id"`
	IsHuman    bool                   `json:"isHuman"`
	IsFinished bool                   `json:"isFinished"`
	Rank       int                    `json:"rank"`
	CardCount  int                    `json:"cardCount"`
	PassesUsed int                    `json:"passesUsed"`
	MaxPasses  int                    `json:"maxPasses"`
	Cards      []*SevensWebOutputCard `json:"cards"`
}

// SevensWebOutputAction 7並べのプレイヤー行動記録
type SevensWebOutputAction struct {
	PlayerIdx  int                  `json:"playerIdx"`
	PlayedCard *SevensWebOutputCard `json:"playedCard"` // nil = パス
}

// SevensWebOutput 7並べWebアウトプット
type SevensWebOutput struct {
	Players      []*SevensWebOutputPlayer `json:"players"`
	CurrentTurn  int                      `json:"currentTurn"`
	TableMinVals [5]int                   `json:"tableMinVals"`
	TableMaxVals [5]int                   `json:"tableMaxVals"`
	GameEndFlag  bool                     `json:"gameEndFlag"`
	CpuActions   []*SevensWebOutputAction `json:"cpuActions"`
	HumanAction  *SevensWebOutputAction   `json:"humanAction"`
	Message      string                   `json:"message"`
}

// SevensWebController 7並べWebコントローラークラス
type SevensWebController struct {
	factory  func() usecases.SevensInteractorIF
	sessions map[string]usecases.SevensInteractorIF
	mu       sync.Mutex
}

// NewSevensWebController コンストラクタ
func NewSevensWebController(factory func() usecases.SevensInteractorIF) *SevensWebController {
	return &SevensWebController{
		factory:  factory,
		sessions: make(map[string]usecases.SevensInteractorIF),
	}
}

// getOrCreateSession セッションIDに対応するインタラクターを取得または生成する
func (swc *SevensWebController) getOrCreateSession(sessionId string) usecases.SevensInteractorIF {
	swc.mu.Lock()
	defer swc.mu.Unlock()
	sgi, ok := swc.sessions[sessionId]
	if !ok {
		sgi = swc.factory()
		swc.sessions[sessionId] = sgi
	}
	return sgi
}

// Exec ゲーム実行
func (swc *SevensWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param SevensWebInput
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(swc.newDefaultOutput("param error."))
		return
	}
	switch param.Command {
	case "q", "quit":
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(swc.newDefaultOutput("bye."))
	case "r", "reset":
		swc.writePresenterResponse(w, swc.getOrCreateSession(param.SessionId).Reset())
	case "p", "play":
		swc.writePresenterResponse(w, swc.getOrCreateSession(param.SessionId).Play(param.Index))
	default:
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(swc.newDefaultOutput("Unsupported command."))
	}
}

// writePresenterResponse プレゼンターの出力を再エンコードせず直接書き込む
func (swc *SevensWebController) writePresenterResponse(w rest.ResponseWriter, responseStr string) {
	if responseStr == "" || !json.Valid([]byte(responseStr)) {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(swc.newDefaultOutput("error."))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = w.WriteJson(json.RawMessage(responseStr))
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (swc *SevensWebController) newDefaultOutput(msg string) *SevensWebOutput {
	return &SevensWebOutput{
		Players:    make([]*SevensWebOutputPlayer, 0),
		CpuActions: make([]*SevensWebOutputAction, 0),
		Message:    msg,
	}
}
