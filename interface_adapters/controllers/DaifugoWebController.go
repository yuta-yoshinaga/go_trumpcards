package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// DaifugoWebInput 大富豪Webインプット
type DaifugoWebInput struct {
	Command string `json:"command"`
	Indices []int  `json:"indices"` // 出すカードのインデックス。play コマンド用。空の場合はパス。
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
	Rank       int                     `json:"rank"`
	CardCount  int                     `json:"cardCount"`
	Cards      []*DaifugoWebOutputCard `json:"cards"`
}

// DaifugoWebOutputAction 大富豪のプレイヤー行動記録
type DaifugoWebOutputAction struct {
	PlayerIdx   int                     `json:"playerIdx"`
	PlayedCards []*DaifugoWebOutputCard `json:"playedCards"` // nil = パス
}

// DaifugoWebOutput 大富豪Webアウトプット
type DaifugoWebOutput struct {
	Players           []*DaifugoWebOutputPlayer `json:"players"`
	CurrentTurn       int                       `json:"currentTurn"`
	TableCards        []*DaifugoWebOutputCard   `json:"tableCards"`
	LastPlayPlayerIdx int                       `json:"lastPlayPlayerIdx"`
	GameEndFlag       bool                      `json:"gameEndFlag"`
	CpuActions        []*DaifugoWebOutputAction `json:"cpuActions"`
	HumanAction       *DaifugoWebOutputAction   `json:"humanAction"`
	Message           string                    `json:"message"`
}

// DaifugoWebController 大富豪Webコントローラークラス
type DaifugoWebController struct {
	dgi usecases.DaifugoInteractorIF
}

// NewDaifugoWebController コンストラクタ
func NewDaifugoWebController(dgi usecases.DaifugoInteractorIF) *DaifugoWebController {
	return &DaifugoWebController{dgi: dgi}
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
			responseStr = dwc.dgi.Reset()
		case "p", "play":
			indices := param.Indices
			if indices == nil {
				indices = []int{}
			}
			responseStr = dwc.dgi.Play(indices)
		default:
			responseStr = `{"message":"Unsupported command."}`
		}
	}
	response := new(DaifugoWebOutput)
	response.Players = make([]*DaifugoWebOutputPlayer, 0)
	response.TableCards = make([]*DaifugoWebOutputCard, 0)
	response.CpuActions = make([]*DaifugoWebOutputAction, 0)
	err = json.Unmarshal([]byte(responseStr), &response)
	if err != nil || responseStr == "" {
		status = http.StatusBadRequest
		response.Message = "error."
	}
	w.WriteHeader(status)
	_ = w.WriteJson(response)
}
