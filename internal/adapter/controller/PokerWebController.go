package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// PokerWebInput ポーカーWebインプット
type PokerWebInput struct {
	Command    string `json:"command"`
	Indices    []int  `json:"indices,omitempty"`
	Amount     int    `json:"amount,omitempty"`
	SessionId  string `json:"sessionId"`
	CpuCount   *int   `json:"cpuCount,omitempty"`
	JokerCount *int   `json:"jokerCount,omitempty"`
}

// GetCommand returns the command string.
func (i PokerWebInput) GetCommand() string { return i.Command }

// GetSessionID returns the session ID string.
func (i PokerWebInput) GetSessionID() string { return i.SessionId }

// PokerWebOutputPlayer ポーカーWebアウトプットプレイヤー
type PokerWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	Cards         []*WebOutputCard `json:"cards"`
	Chips         int              `json:"chips"`
	CurrentBet    int              `json:"currentBet"`
	Folded        bool             `json:"folded"`
	AllIn         bool             `json:"allIn"`
	HandRank      int              `json:"handRank"`
	HandName      string           `json:"handName"`
	ExchangeCount int              `json:"exchangeCount"`
	PlayStyleName string           `json:"playStyleName"`
}

// PokerWebOutputCpuAction ポーカーCPU行動記録
type PokerWebOutputCpuAction struct {
	PlayerIdx int `json:"playerIdx"`
	Action    int `json:"action"`
	Amount    int `json:"amount"`
}

// PokerWebOutputCpuExchange ポーカーCPU交換記録
type PokerWebOutputCpuExchange struct {
	PlayerIdx     int `json:"playerIdx"`
	ExchangeCount int `json:"exchangeCount"`
}

// PokerWebOutputResult ポーカーショーダウン結果
type PokerWebOutputResult struct {
	PlayerIdx int    `json:"playerIdx"`
	HandRank  int    `json:"handRank"`
	HandName  string `json:"handName"`
	WonAmount int    `json:"wonAmount"`
}

// PokerWebOutputSidePot ポーカーサイドポット
type PokerWebOutputSidePot struct {
	Amount          int   `json:"amount"`
	EligiblePlayers []int `json:"eligiblePlayers"`
}

// PokerWebOutputOdds ポーカードローオッズ
type PokerWebOutputOdds struct {
	HandRank    int     `json:"handRank"`
	HandName    string  `json:"handName"`
	Probability float64 `json:"probability"`
	Count       int     `json:"count"`
	Total       int     `json:"total"`
}

// PokerWebOutput ポーカーWebアウトプット
type PokerWebOutput struct {
	Players      []*PokerWebOutputPlayer      `json:"players"`
	Pot          int                          `json:"pot"`
	SidePots     []*PokerWebOutputSidePot     `json:"sidePots"`
	DealerIdx    int                          `json:"dealerIdx"`
	CurrentTurn  int                          `json:"currentTurn"`
	Phase        int                          `json:"phase"`
	GameEndFlag  bool                         `json:"gameEndFlag"`
	LastBet      int                          `json:"lastBet"`
	MinRaise     int                          `json:"minRaise"`
	Ante         int                          `json:"ante"`
	JokerCount   int                          `json:"jokerCount"`
	RoundResults []*PokerWebOutputResult      `json:"roundResults"`
	CpuActions   []*PokerWebOutputCpuAction   `json:"cpuActions"`
	CpuExchanges []*PokerWebOutputCpuExchange `json:"cpuExchanges"`
	Odds         []*PokerWebOutputOdds        `json:"odds,omitempty"`
	Message      string                       `json:"message"`
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
				cfg := domain.DefaultPokerConfig()
				if param.CpuCount != nil {
					cc := *param.CpuCount
					if cc < 1 {
						cc = 1
					} else if cc > 3 {
						cc = 3
					}
					cfg.CpuCount = cc
				}
				if param.JokerCount != nil {
					jc := *param.JokerCount
					if jc < 0 {
						jc = 0
					} else if jc > 2 {
						jc = 2
					}
					cfg.JokerCount = jc
				}
				pwc.writePresenterResponse(w, pi.ResetWithConfig(cfg))
			case "e", "exchange":
				indices := param.Indices
				if indices == nil {
					indices = []int{}
				}
				pwc.writePresenterResponse(w, pi.Exchange(indices))
			case "s", "stand":
				pwc.writePresenterResponse(w, pi.Stand())
			case "f", "fold":
				pwc.writePresenterResponse(w, pi.Action(domain.PokerActionFold, 0))
			case "ck", "check":
				pwc.writePresenterResponse(w, pi.Action(domain.PokerActionCheck, 0))
			case "c", "call":
				pwc.writePresenterResponse(w, pi.Action(domain.PokerActionCall, 0))
			case "b", "bet":
				pwc.writePresenterResponse(w, pi.Action(domain.PokerActionBet, param.Amount))
			case "ra", "raise":
				pwc.writePresenterResponse(w, pi.Action(domain.PokerActionRaise, param.Amount))
			case "a", "allin":
				pwc.writePresenterResponse(w, pi.Action(domain.PokerActionAllIn, 0))
			case "o", "odds":
				indices := param.Indices
				if indices == nil {
					indices = []int{}
				}
				pwc.writePresenterResponse(w, pi.Odds(indices))
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
		Players:      make([]*PokerWebOutputPlayer, 0),
		SidePots:     make([]*PokerWebOutputSidePot, 0),
		RoundResults: make([]*PokerWebOutputResult, 0),
		CpuActions:   make([]*PokerWebOutputCpuAction, 0),
		CpuExchanges: make([]*PokerWebOutputCpuExchange, 0),
		Message:      msg,
	}
}
