package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// OldMaidWebInput ババ抜きWebインプット
type OldMaidWebInput struct {
	Command string `json:"command"`
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

// OldMaidWebOutput ババ抜きWebアウトプット
type OldMaidWebOutput struct {
	Players           []*OldMaidWebOutputPlayer `json:"players"`
	CurrentTurn       int                       `json:"currentTurn"`
	NextDrawTargetIdx int                       `json:"nextDrawTargetIdx"`
	GameEndFlag       bool                      `json:"gameEndFlag"`
	LoserIdx          int                       `json:"loserIdx"`
	LastDrawPlayerIdx int                       `json:"lastDrawPlayerIdx"`
	LastDrawFromIdx   int                       `json:"lastDrawFromIdx"`
	LastDiscardedPairs int                      `json:"lastDiscardedPairs"`
	HasDrawn          bool                      `json:"hasDrawn"`
	Message           string                    `json:"message"`
}

// OldMaidWebController ババ抜きWebコントローラークラス
type OldMaidWebController struct {
	omi usecases.OldMaidInteractorIF
}

// NewOldMaidWebController コンストラクタ
func NewOldMaidWebController(omi usecases.OldMaidInteractorIF) *OldMaidWebController {
	return &OldMaidWebController{omi: omi}
}

// Exec ゲーム実行
func (owc *OldMaidWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param OldMaidWebInput
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
			responseStr = owc.omi.Reset()
		case "d", "draw":
			responseStr = owc.omi.Draw()
		default:
			responseStr = `{"message":"Unsupported command."}`
		}
	}
	response := new(OldMaidWebOutput)
	response.Players = make([]*OldMaidWebOutputPlayer, 0)
	err = json.Unmarshal([]byte(responseStr), &response)
	if err != nil || responseStr == "" {
		status = http.StatusBadRequest
		response.Message = "error."
	}
	w.WriteHeader(status)
	_ = w.WriteJson(response)
}
