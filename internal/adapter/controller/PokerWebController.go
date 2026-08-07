//go:build !js || !wasm || casino

package controller

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"net/http"
)

// PokerWebInput ポーカーWebインプット
type PokerWebInput struct {
	BaseWebInput
	Indices      []int           `json:"indices,omitempty"`
	Amount       int             `json:"amount,omitempty"`
	HumanPlayMs  int             `json:"humanPlayMs,omitempty"`
	CpuCount     *int            `json:"cpuCount,omitempty"`
	JokerCount   *int            `json:"jokerCount,omitempty"`
	BettingLimit *int            `json:"bettingLimit,omitempty"`
	IsLowball    *bool           `json:"isLowball,omitempty"`
	CpuMetaAI    bool            `json:"cpuMetaAI,omitempty"`
	Profile      json.RawMessage `json:"profile,omitempty"`
}

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
	Kickers   string `json:"kickers"`
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

// PokerWebOutputMetaAI メタAI情報
type PokerWebOutputMetaAI struct {
	Enabled        bool    `json:"enabled"`
	GamesPlayed    int     `json:"gamesPlayed"`
	BluffRate      float64 `json:"bluffRate"`
	FoldRate       float64 `json:"foldRate"`
	HesitationMean float64 `json:"hesitationMean"`
}

// PokerWebOutput ポーカーWebアウトプット
type PokerWebOutput struct {
	Players     []*PokerWebOutputPlayer  `json:"players"`
	Pot         int                      `json:"pot"`
	SidePots    []*PokerWebOutputSidePot `json:"sidePots"`
	DealerIdx   int                      `json:"dealerIdx"`
	CurrentTurn int                      `json:"currentTurn"`
	Phase       int                      `json:"phase"`
	// Equity / PotOdds は 2 巡目ベットの判断材料。Holdem 系は EquityDisplay で
	// 出しているのに、5 カードドローには仕組み自体が無かった (#4678)。
	// ベッティングフェーズ以外では省略される。
	Equity       *HoldemWebOutputEquity          `json:"equity,omitempty"`
	PotOdds      *float64                        `json:"potOdds,omitempty"`
	GameEndFlag  bool                            `json:"gameEndFlag"`
	LastBet      int                             `json:"lastBet"`
	MinRaise     int                             `json:"minRaise"`
	Ante         int                             `json:"ante"`
	JokerCount   int                             `json:"jokerCount"`
	BettingLimit int                             `json:"bettingLimit"`
	RaiseCount   int                             `json:"raiseCount"`
	MaxBetAmount int                             `json:"maxBetAmount"`
	RoundResults []*PokerWebOutputResult         `json:"roundResults"`
	CpuActions   []*PokerWebOutputCpuAction      `json:"cpuActions"`
	CpuExchanges []*PokerWebOutputCpuExchange    `json:"cpuExchanges"`
	Odds         []*PokerWebOutputOdds           `json:"odds,omitempty"`
	IsLowball    bool                            `json:"isLowball"`
	MetaAI       *PokerWebOutputMetaAI           `json:"metaAI,omitempty"`
	Profile      *domain.BettingHumanProfileData `json:"profile,omitempty"`
	WebOutputBase
}

// ToConfig builds a PokerConfig from the web input, applying bounds clamping.
func (p PokerWebInput) ToConfig() domain.PokerConfig {
	cfg := domain.DefaultPokerConfig()
	cfg.CpuCount = webutil.BoundedIntPtr(p.CpuCount, domain.PokerCpuCountMin, domain.PokerCpuCountMax, cfg.CpuCount)
	cfg.JokerCount = webutil.BoundedIntPtr(p.JokerCount, 0, domain.PokerJokerCountMax, cfg.JokerCount)
	cfg.BettingLimit = domain.BettingLimitType(webutil.BoundedIntPtr(p.BettingLimit, int(domain.BettingLimitFixed), int(domain.BettingLimitNoLimit), int(cfg.BettingLimit)))
	if p.IsLowball != nil {
		cfg.IsLowball = *p.IsLowball
	}
	cfg.CpuMetaAI = p.CpuMetaAI
	return cfg
}

// PokerWebController ポーカーWebコントローラークラス
type PokerWebController = GameWebController[usecase.PokerInteractorIF, PokerWebInput, *PokerWebOutput]

// NewPokerWebController and NewPokerWebControllerWithProvider are
// the standard and provider-backed constructors for PokerWebController.
var NewPokerWebController, NewPokerWebControllerWithProvider = webControllerPair[usecase.PokerInteractorIF, PokerWebInput, *PokerWebOutput](
	newPokerDefaultOutput, pokerDispatch,
)

func newPokerDefaultOutput(msg string) *PokerWebOutput {
	return &PokerWebOutput{
		Players:       make([]*PokerWebOutputPlayer, 0),
		SidePots:      make([]*PokerWebOutputSidePot, 0),
		RoundResults:  make([]*PokerWebOutputResult, 0),
		CpuActions:    make([]*PokerWebOutputCpuAction, 0),
		CpuExchanges:  make([]*PokerWebOutputCpuExchange, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func pokerDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PokerInteractorIF, param PokerWebInput, _ func(string) *PokerWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.ResetWithConfig(param.ToConfig(), param.Profile))
	case "e", "exchange":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, pi.Exchange(indices))
	case "s", "stand":
		bc.writePresenterResponse(w, pi.Stand())
	case "f", "fold":
		bc.writePresenterResponse(w, pi.Action(domain.PokerActionFold, 0, param.HumanPlayMs))
	case "ck", "check":
		bc.writePresenterResponse(w, pi.Action(domain.PokerActionCheck, 0, param.HumanPlayMs))
	case "c", "call":
		bc.writePresenterResponse(w, pi.Action(domain.PokerActionCall, 0, param.HumanPlayMs))
	case "b", "bet":
		bc.writePresenterResponse(w, pi.Action(domain.PokerActionBet, param.Amount, param.HumanPlayMs))
	case "ra", "raise":
		bc.writePresenterResponse(w, pi.Action(domain.PokerActionRaise, param.Amount, param.HumanPlayMs))
	case "a", "allin":
		bc.writePresenterResponse(w, pi.Action(domain.PokerActionAllIn, 0, param.HumanPlayMs))
	case "o", "odds":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, pi.Odds(indices))
	default:
		return dispatchLog(param.Command, bc, w, pi.ActionLog)
	}
	return true
}
