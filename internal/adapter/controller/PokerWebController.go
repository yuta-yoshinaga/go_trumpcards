package controller

import (
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
		_ = w.WriteJson(pwc.newDefaultOutput("param error."))
		return
	}
	if param.Command == "q" || param.Command == "quit" {
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(pwc.newDefaultOutput("bye."))
		return
	}
	pi, mu, ok := pwc.store.GetWithLock(param.SessionId, pwc.factory)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_ = w.WriteJson(pwc.newDefaultOutput("param error."))
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
		w.WriteHeader(http.StatusOK)
		_ = w.WriteJson(pwc.newDefaultOutput("Unsupported command."))
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
