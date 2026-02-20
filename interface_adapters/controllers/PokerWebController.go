package controllers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// PokerWebInput ポーカーWebインプット
type PokerWebInput struct {
	Command   string `json:"command"`
	Indices   []int  `json:"indices,omitempty"`
	SessionId string `json:"sessionId"`
}

// PokerWebOutputCard ポーカーWebアウトプットカード
type PokerWebOutputCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}

// PokerWebOutputPlayer ポーカーWebアウトプットプレイヤー
type PokerWebOutputPlayer struct {
	HandRank int                   `json:"handRank"`
	HandName string                `json:"handName"`
	Cards    []*PokerWebOutputCard `json:"cards"`
}

// PokerWebOutput ポーカーWebアウトプット
type PokerWebOutput struct {
	Dealer  *PokerWebOutputPlayer `json:"dealer"`
	Player  *PokerWebOutputPlayer `json:"player"`
	Phase   int                   `json:"phase"`
	Message string                `json:"message"`
}

// PokerWebController ポーカーWebコントローラークラス
type PokerWebController struct {
	factory  func() usecases.PokerInteractorIF
	sessions map[string]usecases.PokerInteractorIF
	mu       sync.Mutex
}

// NewPokerWebController コンストラクタ
func NewPokerWebController(factory func() usecases.PokerInteractorIF) *PokerWebController {
	return &PokerWebController{
		factory:  factory,
		sessions: make(map[string]usecases.PokerInteractorIF),
	}
}

// getOrCreateSession セッションIDに対応するインタラクターを取得または生成する
func (pwc *PokerWebController) getOrCreateSession(sessionId string) usecases.PokerInteractorIF {
	pwc.mu.Lock()
	defer pwc.mu.Unlock()
	pi, ok := pwc.sessions[sessionId]
	if !ok {
		pi = pwc.factory()
		pwc.sessions[sessionId] = pi
	}
	return pi
}

// Exec ゲーム実行
func (pwc *PokerWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param PokerWebInput
	status := http.StatusOK
	responseStr := ""
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" {
		status = http.StatusBadRequest
		responseStr = `{"message":"param error."}`
	} else {
		switch param.Command {
		case "q", "quit":
			responseStr = `{"message":"bye."}`
		case "r", "reset":
			responseStr = pwc.getOrCreateSession(param.SessionId).Reset()
		case "e", "exchange":
			indices := param.Indices
			if indices == nil {
				indices = []int{}
			}
			responseStr = pwc.getOrCreateSession(param.SessionId).Exchange(indices)
		case "s", "stand":
			responseStr = pwc.getOrCreateSession(param.SessionId).Stand()
		default:
			responseStr = `{"message":"Unsupported command."}`
		}
	}
	response := new(PokerWebOutput)
	response.Dealer = new(PokerWebOutputPlayer)
	response.Player = new(PokerWebOutputPlayer)
	err = json.Unmarshal([]byte(responseStr), &response)
	if err != nil || responseStr == "" {
		status = http.StatusBadRequest
		response.Message = "error."
	}
	w.WriteHeader(status)
	_ = w.WriteJson(response)
}
