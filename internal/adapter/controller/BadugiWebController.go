//go:build !js || !wasm || casino

package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BadugiWebInput is the request payload for POST /badugi/exec.
type BadugiWebInput struct {
	BaseWebInput
	Indices      []int           `json:"indices,omitempty"`
	Amount       int             `json:"amount,omitempty"`
	HumanPlayMs  int             `json:"humanPlayMs,omitempty"`
	CpuCount     *int            `json:"cpuCount,omitempty"`
	BettingLimit *int            `json:"bettingLimit,omitempty"`
	CpuMetaAI    bool            `json:"cpuMetaAI,omitempty"`
	Profile      json.RawMessage `json:"profile,omitempty"`
}

// BadugiWebOutputPlayer is the per-seat payload in BadugiWebOutput.
type BadugiWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	Cards         []*WebOutputCard `json:"cards"`
	Chips         int              `json:"chips"`
	CurrentBet    int              `json:"currentBet"`
	Folded        bool             `json:"folded"`
	AllIn         bool             `json:"allIn"`
	HandSize      int              `json:"handSize"`
	HandName      string           `json:"handName"`
	DrawCount     int              `json:"drawCount"`
	TotalDraws    int              `json:"totalDraws"`
	PlayStyleName string           `json:"playStyleName"`
	BestCards     []*WebOutputCard `json:"bestCards,omitempty"`
}

// BadugiWebOutputCpuAction is a single CPU betting decision record.
type BadugiWebOutputCpuAction struct {
	PlayerIdx  int    `json:"playerIdx"`
	Action     int    `json:"action"`
	Amount     int    `json:"amount"`
	DrawIndex  int    `json:"drawIndex"`
	RoundLabel string `json:"roundLabel"`
}

// BadugiWebOutputCpuExchange is a single CPU draw decision record.
type BadugiWebOutputCpuExchange struct {
	PlayerIdx     int `json:"playerIdx"`
	DrawIndex     int `json:"drawIndex"`
	ExchangeCount int `json:"exchangeCount"`
}

// BadugiWebOutputResult is a single showdown result.
type BadugiWebOutputResult struct {
	PlayerIdx int    `json:"playerIdx"`
	HandSize  int    `json:"handSize"`
	HandName  string `json:"handName"`
	WonAmount int    `json:"wonAmount"`
}

// BadugiWebOutputSidePot mirrors domain.SidePot over the wire.
type BadugiWebOutputSidePot struct {
	Amount          int   `json:"amount"`
	EligiblePlayers []int `json:"eligiblePlayers"`
}

// BadugiWebOutputMetaAI exposes meta-AI profile stats to the UI.
type BadugiWebOutputMetaAI struct {
	Enabled        bool    `json:"enabled"`
	GamesPlayed    int     `json:"gamesPlayed"`
	BluffRate      float64 `json:"bluffRate"`
	FoldRate       float64 `json:"foldRate"`
	HesitationMean float64 `json:"hesitationMean"`
}

// BadugiWebOutput is the full JSON response for /badugi/exec.
type BadugiWebOutput struct {
	Players      []*BadugiWebOutputPlayer        `json:"players"`
	Pot          int                             `json:"pot"`
	SidePots     []*BadugiWebOutputSidePot       `json:"sidePots"`
	DealerIdx    int                             `json:"dealerIdx"`
	CurrentTurn  int                             `json:"currentTurn"`
	Phase        int                             `json:"phase"`
	DrawIndex    int                             `json:"drawIndex"`
	GameEndFlag  bool                            `json:"gameEndFlag"`
	LastBet      int                             `json:"lastBet"`
	MinRaise     int                             `json:"minRaise"`
	Ante         int                             `json:"ante"`
	BettingLimit int                             `json:"bettingLimit"`
	RaiseCount   int                             `json:"raiseCount"`
	MaxBetAmount int                             `json:"maxBetAmount"`
	RoundResults []*BadugiWebOutputResult        `json:"roundResults"`
	CpuActions   []*BadugiWebOutputCpuAction     `json:"cpuActions"`
	CpuExchanges []*BadugiWebOutputCpuExchange   `json:"cpuExchanges"`
	MetaAI       *BadugiWebOutputMetaAI          `json:"metaAI,omitempty"`
	Profile      *domain.BettingHumanProfileData `json:"profile,omitempty"`
	WebOutputBase
}

// ToConfig builds a BadugiConfig from the web input, clamping numeric fields
// to their domain-valid ranges.
func (p BadugiWebInput) ToConfig() domain.BadugiConfig {
	cfg := domain.DefaultBadugiConfig()
	cfg.CpuCount = webutil.BoundedIntPtr(p.CpuCount, domain.BadugiCpuCountMin, domain.BadugiCpuCountMax, cfg.CpuCount)
	cfg.BettingLimit = domain.BettingLimitType(webutil.BoundedIntPtr(
		p.BettingLimit, int(domain.BettingLimitFixed), int(domain.BettingLimitNoLimit), int(cfg.BettingLimit),
	))
	cfg.CpuMetaAI = p.CpuMetaAI
	return cfg
}

// BadugiWebController is the HTTP handler type for POST /badugi/exec.
type BadugiWebController = GameWebController[usecase.BadugiInteractorIF, BadugiWebInput, *BadugiWebOutput]

// NewBadugiWebController / NewBadugiWebControllerWithProvider are the standard
// and session-provider-backed constructors.
var NewBadugiWebController, NewBadugiWebControllerWithProvider = webControllerPair[usecase.BadugiInteractorIF, BadugiWebInput, *BadugiWebOutput](
	newBadugiDefaultOutput, badugiDispatch,
)

func newBadugiDefaultOutput(msg string) *BadugiWebOutput {
	return &BadugiWebOutput{
		Players:       make([]*BadugiWebOutputPlayer, 0),
		SidePots:      make([]*BadugiWebOutputSidePot, 0),
		RoundResults:  make([]*BadugiWebOutputResult, 0),
		CpuActions:    make([]*BadugiWebOutputCpuAction, 0),
		CpuExchanges:  make([]*BadugiWebOutputCpuExchange, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func badugiDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BadugiInteractorIF, param BadugiWebInput, _ func(string) *BadugiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig(), param.Profile))
	case "e", "exchange":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, bi.Exchange(indices))
	case "s", "stand":
		bc.writePresenterResponse(w, bi.Stand())
	case "f", "fold":
		bc.writePresenterResponse(w, bi.Action(domain.BadugiActionFold, 0, param.HumanPlayMs))
	case "ck", "check":
		bc.writePresenterResponse(w, bi.Action(domain.BadugiActionCheck, 0, param.HumanPlayMs))
	case "c", "call":
		bc.writePresenterResponse(w, bi.Action(domain.BadugiActionCall, 0, param.HumanPlayMs))
	case "b", "bet":
		bc.writePresenterResponse(w, bi.Action(domain.BadugiActionBet, param.Amount, param.HumanPlayMs))
	case "ra", "raise":
		bc.writePresenterResponse(w, bi.Action(domain.BadugiActionRaise, param.Amount, param.HumanPlayMs))
	case "a", "allin":
		bc.writePresenterResponse(w, bi.Action(domain.BadugiActionAllIn, 0, param.HumanPlayMs))
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
