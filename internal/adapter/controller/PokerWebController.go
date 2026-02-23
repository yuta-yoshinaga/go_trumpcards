package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// PokerWebInput ポーカーWebインプット
type PokerWebInput struct {
	Command   string `json:"command"`
	Indices   []int  `json:"indices,omitempty"`
	Amount    int    `json:"amount,omitempty"`
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
	Chips    int                   `json:"chips"`
	Bet      int                   `json:"bet"`
}

// PokerWebOutput ポーカーWebアウトプット
type PokerWebOutput struct {
	Dealer  *PokerWebOutputPlayer `json:"dealer"`
	Player  *PokerWebOutputPlayer `json:"player"`
	Phase   int                   `json:"phase"`
	Message string                `json:"message"`
	Pot     int                   `json:"pot"`
	Ante    int                   `json:"ante"`
}

// PokerWebController ポーカーWebコントローラークラス
type PokerWebController struct {
	factory func() usecase.PokerInteractorIF
	store   *SessionStore[usecase.PokerInteractorIF]
}

// NewPokerWebController コンストラクタ
func NewPokerWebController(factory func() usecase.PokerInteractorIF) *PokerWebController {
	return &PokerWebController{
		factory: factory,
		store:   NewSessionStore[usecase.PokerInteractorIF](),
	}
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
	} else if param.Command == "q" || param.Command == "quit" {
		responseStr = `{"message":"bye."}`
	} else {
		pi, mu, ok := pwc.store.GetWithLock(param.SessionId, pwc.factory)
		if !ok {
			status = http.StatusBadRequest
			responseStr = `{"message":"param error."}`
		} else {
			mu.Lock()
			defer mu.Unlock()
			switch param.Command {
			case "r", "reset":
				responseStr = pi.Reset()
			case "e", "exchange":
				indices := param.Indices
				if indices == nil {
					indices = []int{}
				}
				responseStr = pi.Exchange(indices)
			case "s", "stand":
				responseStr = pi.Stand()
			case "b", "bet":
				responseStr = pi.Bet(param.Amount)
			case "c", "call":
				responseStr = pi.Call()
			case "ra", "raise":
				responseStr = pi.Raise(param.Amount)
			case "f", "fold":
				responseStr = pi.Fold()
			case "ck", "check":
				responseStr = pi.Check()
			default:
				responseStr = `{"message":"Unsupported command."}`
			}
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
