//go:build !js || !wasm || casino

package controller

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"net/http"
)

// IndianPokerWebInput インディアンポーカーWebインプット
type IndianPokerWebInput struct {
	BaseWebInput
	Amount       int             `json:"amount,omitempty"`
	HumanPlayMs  int             `json:"humanPlayMs,omitempty"`
	Ante         *int            `json:"ante,omitempty"`
	BettingLimit *int            `json:"bettingLimit,omitempty"`
	CpuMetaAI    *bool           `json:"cpuMetaAI,omitempty"`
	Profile      json.RawMessage `json:"profile,omitempty"`
}

// IndianPokerWebOutputPlayer インディアンポーカーWebアウトプットプレイヤー
type IndianPokerWebOutputPlayer struct {
	ID            int            `json:"id"`
	IsHuman       bool           `json:"isHuman"`
	Card          *WebOutputCard `json:"card"`
	Chips         int            `json:"chips"`
	CurrentBet    int            `json:"currentBet"`
	Folded        bool           `json:"folded"`
	AllIn         bool           `json:"allIn"`
	PlayStyleName string         `json:"playStyleName"`
}

// IndianPokerWebOutputCpuAction インディアンポーカーCPU行動記録
type IndianPokerWebOutputCpuAction struct {
	PlayerIdx int `json:"playerIdx"`
	Action    int `json:"action"`
	Amount    int `json:"amount"`
}

// IndianPokerWebOutputResult インディアンポーカーショーダウン結果
type IndianPokerWebOutputResult struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
	CardRank  int            `json:"cardRank"`
	WonAmount int            `json:"wonAmount"`
}

// IndianPokerWebOutputSidePot インディアンポーカーサイドポット
type IndianPokerWebOutputSidePot struct {
	Amount          int   `json:"amount"`
	EligiblePlayers []int `json:"eligiblePlayers"`
}

// IndianPokerWebOutputMetaAI メタAI情報
type IndianPokerWebOutputMetaAI struct {
	Enabled        bool    `json:"enabled"`
	GamesPlayed    int     `json:"gamesPlayed"`
	BluffRate      float64 `json:"bluffRate"`
	FoldRate       float64 `json:"foldRate"`
	HesitationMean float64 `json:"hesitationMean"`
}

// IndianPokerWebOutput インディアンポーカーWebアウトプット
type IndianPokerWebOutput struct {
	Players     []*IndianPokerWebOutputPlayer  `json:"players"`
	Pot         int                            `json:"pot"`
	SidePots    []*IndianPokerWebOutputSidePot `json:"sidePots"`
	DealerIdx   int                            `json:"dealerIdx"`
	CurrentTurn int                            `json:"currentTurn"`
	Phase       int                            `json:"phase"`
	// EstimatedStrength は人間プレイヤーの推定勝率 (0-100)。CUI が出しているのと
	// 同じ domain の値。**フロントで計算し直さない** -- 以前は別実装があり、
	// エースの扱いを誤って最も危険な場面ほど勝率を高く見せていた (#4690/#5505)。
	EstimatedStrength int                                 `json:"estimatedStrength"`
	GameEndFlag       bool                                `json:"gameEndFlag"`
	LastBet           int                                 `json:"lastBet"`
	MinRaise          int                                 `json:"minRaise"`
	BettingLimit      int                                 `json:"bettingLimit"`
	RaiseCount        int                                 `json:"raiseCount"`
	MaxBetAmount      int                                 `json:"maxBetAmount"`
	RoundResults      []*IndianPokerWebOutputResult       `json:"roundResults"`
	CpuActions        []*IndianPokerWebOutputCpuAction    `json:"cpuActions"`
	HandCount         int                                 `json:"handCount"`
	Ante              int                                 `json:"ante"`
	MetaAI            *IndianPokerWebOutputMetaAI         `json:"metaAI,omitempty"`
	Profile           *domain.IndianPokerHumanProfileData `json:"profile,omitempty"`
	WebOutputBase
}

// ToConfig builds an IndianPokerConfig from the web input.
func (p IndianPokerWebInput) ToConfig() domain.IndianPokerConfig {
	cfg := domain.DefaultIndianPokerConfig()
	if p.Ante != nil && *p.Ante >= 1 {
		cfg.Ante = *p.Ante
	}
	if p.BettingLimit != nil {
		bl := *p.BettingLimit
		if bl < 0 {
			bl = 0
		} else if bl > 2 {
			bl = 2
		}
		cfg.BettingLimit = domain.BettingLimitType(bl)
	}
	if p.CpuMetaAI != nil {
		cfg.CpuMetaAI = *p.CpuMetaAI
	}
	return cfg
}

// IndianPokerWebController インディアンポーカーWebコントローラークラス
type IndianPokerWebController = GameWebController[usecase.IndianPokerInteractorIF, IndianPokerWebInput, *IndianPokerWebOutput]

// NewIndianPokerWebController and NewIndianPokerWebControllerWithProvider are
// the standard and provider-backed constructors for IndianPokerWebController.
var NewIndianPokerWebController, NewIndianPokerWebControllerWithProvider = webControllerPair[usecase.IndianPokerInteractorIF, IndianPokerWebInput, *IndianPokerWebOutput](
	newIndianPokerDefaultOutput, indianPokerDispatch,
)

func newIndianPokerDefaultOutput(msg string) *IndianPokerWebOutput {
	return &IndianPokerWebOutput{
		Players:       make([]*IndianPokerWebOutputPlayer, 0),
		SidePots:      make([]*IndianPokerWebOutputSidePot, 0),
		RoundResults:  make([]*IndianPokerWebOutputResult, 0),
		CpuActions:    make([]*IndianPokerWebOutputCpuAction, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func indianPokerDispatch(bc *baseController, w http.ResponseWriter, ipi usecase.IndianPokerInteractorIF, param IndianPokerWebInput, newDefault func(string) *IndianPokerWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		cfg := param.ToConfig()
		bc.writePresenterResponse(w, ipi.ResetWithConfig(cfg, param.Profile))
	case "f", "fold":
		bc.writePresenterResponse(w, ipi.Action(domain.IndianPokerActionFold, 0, param.HumanPlayMs))
	case "ck", "check":
		bc.writePresenterResponse(w, ipi.Action(domain.IndianPokerActionCheck, 0, param.HumanPlayMs))
	case "c", "call":
		bc.writePresenterResponse(w, ipi.Action(domain.IndianPokerActionCall, 0, param.HumanPlayMs))
	case "b", "bet":
		bc.writePresenterResponse(w, ipi.Action(domain.IndianPokerActionBet, param.Amount, param.HumanPlayMs))
	case "ra", "raise":
		bc.writePresenterResponse(w, ipi.Action(domain.IndianPokerActionRaise, param.Amount, param.HumanPlayMs))
	case "a", "allin":
		bc.writePresenterResponse(w, ipi.Action(domain.IndianPokerActionAllIn, 0, param.HumanPlayMs))
	default:
		return dispatchLog(param.Command, bc, w, ipi.ActionLog)
	}
	return true
}
