package controllers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// BlackJackWebInput ブラックジャックWebインプット
type BlackJackWebInput struct {
	Command   string `json:"command"`
	SessionId string `json:"sessionId"`
}

// BlackJackWebOutputCard ブラックジャックWebアウトプットカード
type BlackJackWebOutputCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}

// BlackJackWebOutputPlayer ブラックジャックWebアウトプットプレイヤー
type BlackJackWebOutputPlayer struct {
	Score int                       `json:"score"`
	Cards []*BlackJackWebOutputCard `json:"cards"`
}

// BlackJackWebOutput ブラックジャックWebアウトプット
type BlackJackWebOutput struct {
	Dealer  *BlackJackWebOutputPlayer `json:"dealer"`
	Player  *BlackJackWebOutputPlayer `json:"player"`
	Message string                    `json:"message"`
}

// BlackJackWebController ブラックジャックWebコントローラークラス
type BlackJackWebController struct {
	factory  func() usecases.BlackJackInteractorIF
	sessions map[string]usecases.BlackJackInteractorIF
	mu       sync.Mutex
}

// NewBlackJackWebController コンストラクタ
func NewBlackJackWebController(factory func() usecases.BlackJackInteractorIF) *BlackJackWebController {
	return &BlackJackWebController{
		factory:  factory,
		sessions: make(map[string]usecases.BlackJackInteractorIF),
	}
}

// getOrCreateSession セッションIDに対応するインタラクターを取得または生成する
func (bwc *BlackJackWebController) getOrCreateSession(sessionId string) usecases.BlackJackInteractorIF {
	bwc.mu.Lock()
	defer bwc.mu.Unlock()
	bji, ok := bwc.sessions[sessionId]
	if !ok {
		bji = bwc.factory()
		bwc.sessions[sessionId] = bji
	}
	return bji
}

// Exec ゲーム実行
func (bwc *BlackJackWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param BlackJackWebInput
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
			responseStr = bwc.getOrCreateSession(param.SessionId).Reset()
		case "h", "hit":
			responseStr = bwc.getOrCreateSession(param.SessionId).Hit()
		case "s", "stand":
			responseStr = bwc.getOrCreateSession(param.SessionId).Stand()
		default:
			responseStr = `{"message":"Unsupported command."}`
		}
	}
	response := new(BlackJackWebOutput)
	response.Dealer = new(BlackJackWebOutputPlayer)
	response.Player = new(BlackJackWebOutputPlayer)
	err = json.Unmarshal([]byte(responseStr), &response)
	if err != nil || responseStr == "" {
		status = http.StatusBadRequest
		response.Message = "error."
	}
	w.WriteHeader(status)
	_ = w.WriteJson(response)
}
