package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// PokerWebInput ポーカーWebインプット
type PokerWebInput struct {
	Command string `json:"command"`
	Indices []int  `json:"indices,omitempty"`
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
	pi usecases.PokerInteractorIF
}

// NewPokerWebController コンストラクタ
func NewPokerWebController(pi usecases.PokerInteractorIF) *PokerWebController {
	return &PokerWebController{
		pi: pi,
	}
}

// Exec ゲーム実行
func (pwc *PokerWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param PokerWebInput
	status := http.StatusOK
	responseStr := ""
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" {
		status = http.StatusBadRequest
		responseStr = `{"message":"param error."}`
	} else {
		switch param.Command {
		case "q", "quit":
			responseStr = `{"message":"bye."}`
		case "r", "reset":
			responseStr = pwc.pi.Reset()
		case "e", "exchange":
			indices := param.Indices
			if indices == nil {
				indices = []int{}
			}
			responseStr = pwc.pi.Exchange(indices)
		case "s", "stand":
			responseStr = pwc.pi.Stand()
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
