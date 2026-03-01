package controller

import (
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

// GetCommand returns the command string.
func (i PokerWebInput) GetCommand() string { return i.Command }

// GetSessionID returns the session ID string.
func (i PokerWebInput) GetSessionID() string { return i.SessionId }

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
	execWithSession(&pwc.baseController, w, r, pwc.store, pwc.factory,
		func(msg string) any { return pwc.newDefaultOutput(msg) },
		nil,
		func(w rest.ResponseWriter, pi usecase.PokerInteractorIF, param PokerWebInput) bool {
			switch param.Command {
			case "r", "reset":
				pwc.writePresenterResponse(w, pi.Reset())
			case "e", "exchange":
				indices := param.Indices
				if indices == nil {
					indices = []int{}
				}
				pwc.writePresenterResponse(w, pi.Exchange(indices))
			case "s", "stand":
				pwc.writePresenterResponse(w, pi.Stand())
			case "b", "bet":
				pwc.writePresenterResponse(w, pi.Bet(param.Amount))
			case "c", "call":
				pwc.writePresenterResponse(w, pi.Call())
			case "ra", "raise":
				pwc.writePresenterResponse(w, pi.Raise(param.Amount))
			case "f", "fold":
				pwc.writePresenterResponse(w, pi.Fold())
			case "ck", "check":
				pwc.writePresenterResponse(w, pi.Check())
			default:
				return false
			}
			return true
		})
}

// Stop stops the background cleanup goroutine of the session store.
func (pwc *PokerWebController) Stop() {
	pwc.store.Stop()
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (pwc *PokerWebController) newDefaultOutput(msg string) *PokerWebOutput {
	return &PokerWebOutput{
		Dealer:  &PokerWebOutputPlayer{},
		Player:  &PokerWebOutputPlayer{},
		Message: msg,
	}
}
