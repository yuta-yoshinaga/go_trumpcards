package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
)

// BlackJackWebInput ブラックジャックWebインプット
type BlackJackWebInput struct {
	Command   string `json:"command"`
	Amount    int    `json:"amount,omitempty"`
	SessionId string `json:"sessionId"`
}

// BlackJackWebOutputCard ブラックジャックWebアウトプットカード
type BlackJackWebOutputCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}

// BlackJackWebOutputHand ブラックジャックWebアウトプットハンド
type BlackJackWebOutputHand struct {
	Score       int                       `json:"score"`
	Cards       []*BlackJackWebOutputCard `json:"cards"`
	Bet         int                       `json:"bet"`
	Stood       bool                      `json:"stood"`
	Doubled     bool                      `json:"doubled"`
	Busted      bool                      `json:"busted"`
	IsBlackJack bool                      `json:"isBlackJack"`
	CanSplit    bool                      `json:"canSplit"`
}

// BlackJackWebOutputPlayer ブラックジャックWebアウトプットプレイヤー
type BlackJackWebOutputPlayer struct {
	Score int                       `json:"score"`
	Cards []*BlackJackWebOutputCard `json:"cards"`
	Chips int                       `json:"chips"`
}

// BlackJackWebOutput ブラックジャックWebアウトプット
type BlackJackWebOutput struct {
	Dealer             *BlackJackWebOutputPlayer `json:"dealer"`
	Player             *BlackJackWebOutputPlayer `json:"player"`
	Hands              []*BlackJackWebOutputHand `json:"hands,omitempty"`
	CurrentHandIdx     int                       `json:"currentHandIdx"`
	Phase              int                       `json:"phase"`
	InsuranceBet       int                       `json:"insuranceBet"`
	InsuranceAvailable bool                      `json:"insuranceAvailable"`
	Message            string                    `json:"message"`
}

// BlackJackWebController ブラックジャックWebコントローラークラス
type BlackJackWebController struct {
	factory func() usecases.BlackJackInteractorIF
	store   *SessionStore[usecases.BlackJackInteractorIF]
}

// NewBlackJackWebController コンストラクタ
func NewBlackJackWebController(factory func() usecases.BlackJackInteractorIF) *BlackJackWebController {
	return &BlackJackWebController{
		factory: factory,
		store:   NewSessionStore[usecases.BlackJackInteractorIF](),
	}
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
	} else if param.Command == "q" || param.Command == "quit" {
		responseStr = `{"message":"bye."}`
	} else {
		bji, mu, ok := bwc.store.GetWithLock(param.SessionId, bwc.factory)
		if !ok {
			status = http.StatusBadRequest
			responseStr = `{"message":"param error."}`
		} else {
			mu.Lock()
			switch param.Command {
			case "r", "reset":
				responseStr = bji.Reset()
			case "h", "hit":
				responseStr = bji.Hit()
			case "s", "stand":
				responseStr = bji.Stand()
			case "b", "bet":
				responseStr = bji.Bet(param.Amount)
			case "d", "doubledown":
				responseStr = bji.DoubleDown()
			case "sp", "split":
				responseStr = bji.Split()
			case "i", "insurance":
				responseStr = bji.Insurance()
			case "di", "declineinsurance":
				responseStr = bji.DeclineInsurance()
			default:
				responseStr = `{"message":"Unsupported command."}`
			}
			mu.Unlock()
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
