package controllers

import (
	"encoding/json"
	"net/http"
	"go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// BlackJackWebInput ブラックジャックWebインプット
type BlackJackWebInput struct {
	Command string `json:"command"`
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
	bji usecases.BlackJackInteractorIF
}

// NewBlackJackWebController コンストラクタ
func NewBlackJackWebController(bji usecases.BlackJackInteractorIF) *BlackJackWebController {
	return &BlackJackWebController{
		bji: bji,
	}
}

// Exec ゲーム実行
func (bwc *BlackJackWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param BlackJackWebInput
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
			responseStr = bwc.bji.Reset()
		case "h", "hit":
			responseStr = bwc.bji.Hit()
		case "s", "stand":
			responseStr = bwc.bji.Stand()
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
