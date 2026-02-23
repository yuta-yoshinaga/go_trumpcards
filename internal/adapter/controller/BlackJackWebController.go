package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

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
	factory func() usecase.BlackJackInteractorIF
	store   *SessionStore[usecase.BlackJackInteractorIF]
}

// NewBlackJackWebController コンストラクタ
func NewBlackJackWebController(factory func() usecase.BlackJackInteractorIF) *BlackJackWebController {
	return &BlackJackWebController{
		factory: factory,
		store:   NewSessionStore[usecase.BlackJackInteractorIF](),
	}
}

// Exec ゲーム実行
func (bwc *BlackJackWebController) Exec(w rest.ResponseWriter, r *rest.Request) {
	var param BlackJackWebInput
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(bwc.newDefaultOutput("param error."))
		return
	}
	if param.Command == "q" || param.Command == "quit" {
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(bwc.newDefaultOutput("bye."))
		return
	}
	bji, mu, ok := bwc.store.GetWithLock(param.SessionId, bwc.factory)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(bwc.newDefaultOutput("param error."))
		return
	}
	mu.Lock()
	defer mu.Unlock()
	switch param.Command {
	case "r", "reset":
		bwc.writePresenterResponse(w, bji.Reset())
	case "h", "hit":
		bwc.writePresenterResponse(w, bji.Hit())
	case "s", "stand":
		bwc.writePresenterResponse(w, bji.Stand())
	case "b", "bet":
		bwc.writePresenterResponse(w, bji.Bet(param.Amount))
	case "d", "doubledown":
		bwc.writePresenterResponse(w, bji.DoubleDown())
	case "sp", "split":
		bwc.writePresenterResponse(w, bji.Split())
	case "i", "insurance":
		bwc.writePresenterResponse(w, bji.Insurance())
	case "di", "declineinsurance":
		bwc.writePresenterResponse(w, bji.DeclineInsurance())
	default:
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(bwc.newDefaultOutput("Unsupported command."))
	}
}

// writePresenterResponse プレゼンターの出力を再エンコードせず直接書き込む
func (bwc *BlackJackWebController) writePresenterResponse(w rest.ResponseWriter, responseStr string) {
	if responseStr == "" || !json.Valid([]byte(responseStr)) {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(bwc.newDefaultOutput("error."))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = w.WriteJson(json.RawMessage(responseStr))
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (bwc *BlackJackWebController) newDefaultOutput(msg string) *BlackJackWebOutput {
	return &BlackJackWebOutput{
		Dealer:  &BlackJackWebOutputPlayer{},
		Player:  &BlackJackWebOutputPlayer{},
		Message: msg,
	}
}
