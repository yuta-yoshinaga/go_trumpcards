package controller

import (
	"log"
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

// PokerWebOutputPlayer ポーカーWebアウトプットプレイヤー
type PokerWebOutputPlayer struct {
	HandRank int              `json:"handRank"`
	HandName string           `json:"handName"`
	Cards    []*WebOutputCard `json:"cards"`
	Chips    int              `json:"chips"`
	Bet      int              `json:"bet"`
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
	baseController
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
	err := r.DecodeJsonPayload(&param)
	if err != nil || param.Command == "" || param.SessionId == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(pwc.newDefaultOutput("param error.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	if param.Command == "q" || param.Command == "quit" {
		w.WriteHeader(http.StatusOK)
		if err := w.WriteJson(pwc.newDefaultOutput("bye.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	pi, mu, ok := pwc.store.GetWithLock(param.SessionId, pwc.factory)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(pwc.newDefaultOutput("param error.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
		return
	}
	mu.Lock()
	defer mu.Unlock()
	errOutput := pwc.newDefaultOutput("error.")
	switch param.Command {
	case "r", "reset":
		pwc.writePresenterResponse(w, pi.Reset(), errOutput)
	case "e", "exchange":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		pwc.writePresenterResponse(w, pi.Exchange(indices), errOutput)
	case "s", "stand":
		pwc.writePresenterResponse(w, pi.Stand(), errOutput)
	case "b", "bet":
		pwc.writePresenterResponse(w, pi.Bet(param.Amount), errOutput)
	case "c", "call":
		pwc.writePresenterResponse(w, pi.Call(), errOutput)
	case "ra", "raise":
		pwc.writePresenterResponse(w, pi.Raise(param.Amount), errOutput)
	case "f", "fold":
		pwc.writePresenterResponse(w, pi.Fold(), errOutput)
	case "ck", "check":
		pwc.writePresenterResponse(w, pi.Check(), errOutput)
	default:
		w.WriteHeader(http.StatusBadRequest)
		if err := w.WriteJson(pwc.newDefaultOutput("Unsupported command.")); err != nil {
			log.Printf("WriteJson error: %v", err)
		}
	}
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (pwc *PokerWebController) newDefaultOutput(msg string) *PokerWebOutput {
	return &PokerWebOutput{
		Dealer:  &PokerWebOutputPlayer{},
		Player:  &PokerWebOutputPlayer{},
		Message: msg,
	}
}
