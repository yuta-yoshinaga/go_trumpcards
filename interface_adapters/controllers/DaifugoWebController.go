package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// DaifugoWebInput 大富豪Webインプット
type DaifugoWebInput struct {
	Command     string `json:"command"`
	CardIndices []int  `json:"cardIndices"`
}

// DaifugoWebOutputCard 大富豪Webアウトプットカード
type DaifugoWebOutputCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}

// DaifugoWebOutputPlayer 大富豪Webアウトプットプレイヤー
type DaifugoWebOutputPlayer struct {
	ID         int                     `json:"id"`
	IsHuman    bool                    `json:"isHuman"`
	IsFinished bool                    `json:"isFinished"`
	CardCount  int                     `json:"cardCount"`
	Cards      []*DaifugoWebOutputCard `json:"cards"`
	Rank       int                     `json:"rank"` // 0:大富豪, ..., 3:大貧民
}

// DaifugoWebOutput 大富豪Webアウトプット
type DaifugoWebOutput struct {
	Players      []*DaifugoWebOutputPlayer `json:"players"`
	CurrentTurn  int                       `json:"currentTurn"`
	LastPlay     []*DaifugoWebOutputCard   `json:"lastPlay"`
	IsRevolution bool                      `json:"isRevolution"`
	GameEndFlag  bool                      `json:"gameEndFlag"`
	Message      string                    `json:"message"`
}

// DaifugoWebController 大富豪Webコントローラークラス
type DaifugoWebController struct {
	di usecases.DaifugoInteractorIF
}

// NewDaifugoWebController コンストラクタ
func NewDaifugoWebController(di usecases.DaifugoInteractorIF) *DaifugoWebController {
	return &DaifugoWebController{di: di}
}

// Exec ゲーム実行
func (dwc *DaifugoWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param DaifugoWebInput
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
			responseStr = dwc.di.Reset()
		case "p", "play":
			responseStr = dwc.di.Play(param.CardIndices)
		case "s", "pass":
			responseStr = dwc.di.Pass()
		default:
			responseStr = `{"message":"Unsupported command."}`
		}
	}
	response := new(DaifugoWebOutput)
	response.Players = make([]*DaifugoWebOutputPlayer, 0)
	response.LastPlay = make([]*DaifugoWebOutputCard, 0)
	err = json.Unmarshal([]byte(responseStr), &response)
	if err != nil || responseStr == "" {
		status = http.StatusBadRequest
		response.Message = "error."
	}
	w.WriteHeader(status)
	_ = w.WriteJson(response)
}
