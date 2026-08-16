//go:build !js || !wasm || casino

package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DeuceToSevenWebInput is the request payload for POST /deucetoseven/exec.
type DeuceToSevenWebInput struct {
	BaseWebInput
	Indices      []int           `json:"indices,omitempty"`
	Amount       int             `json:"amount,omitempty"`
	HumanPlayMs  int             `json:"humanPlayMs,omitempty"`
	CpuCount     *int            `json:"cpuCount,omitempty"`
	BettingLimit *int            `json:"bettingLimit,omitempty"`
	CpuMetaAI    bool            `json:"cpuMetaAI,omitempty"`
	Profile      json.RawMessage `json:"profile,omitempty"`
}

// DeuceToSevenWebOutputPlayer is the per-seat payload in DeuceToSevenWebOutput.
type DeuceToSevenWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	Cards         []*WebOutputCard `json:"cards"`
	Chips         int              `json:"chips"`
	CurrentBet    int              `json:"currentBet"`
	Folded        bool             `json:"folded"`
	AllIn         bool             `json:"allIn"`
	HandRank      int              `json:"handRank"`
	HandName      string           `json:"handName"`
	DrawCount     int              `json:"drawCount"`
	TotalDraws    int              `json:"totalDraws"`
	PlayStyleName string           `json:"playStyleName"`
}

// DeuceToSevenWebOutputCpuAction is a single CPU betting decision record.
type DeuceToSevenWebOutputCpuAction struct {
	PlayerIdx  int    `json:"playerIdx"`
	Action     int    `json:"action"`
	Amount     int    `json:"amount"`
	DrawIndex  int    `json:"drawIndex"`
	RoundLabel string `json:"roundLabel"`
}

// DeuceToSevenWebOutputCpuExchange is a single CPU draw decision record.
type DeuceToSevenWebOutputCpuExchange struct {
	PlayerIdx     int `json:"playerIdx"`
	DrawIndex     int `json:"drawIndex"`
	ExchangeCount int `json:"exchangeCount"`
}

// DeuceToSevenWebOutputResult is a single showdown result.
type DeuceToSevenWebOutputResult struct {
	PlayerIdx int    `json:"playerIdx"`
	HandRank  int    `json:"handRank"`
	HandName  string `json:"handName"`
	WonAmount int    `json:"wonAmount"`
}

// DeuceToSevenWebOutputSidePot mirrors domain.SidePot over the wire.
type DeuceToSevenWebOutputSidePot struct {
	Amount          int   `json:"amount"`
	EligiblePlayers []int `json:"eligiblePlayers"`
}

// DeuceToSevenWebOutputMetaAI exposes meta-AI profile stats to the UI.
type DeuceToSevenWebOutputMetaAI struct {
	Enabled        bool    `json:"enabled"`
	GamesPlayed    int     `json:"gamesPlayed"`
	BluffRate      float64 `json:"bluffRate"`
	FoldRate       float64 `json:"foldRate"`
	HesitationMean float64 `json:"hesitationMean"`
}

// DeuceToSevenWebOutput is the full JSON response for /deucetoseven/exec.
type DeuceToSevenWebOutput struct {
	Players      []*DeuceToSevenWebOutputPlayer      `json:"players"`
	Pot          int                                 `json:"pot"`
	SidePots     []*DeuceToSevenWebOutputSidePot     `json:"sidePots"`
	DealerIdx    int                                 `json:"dealerIdx"`
	CurrentTurn  int                                 `json:"currentTurn"`
	Phase        int                                 `json:"phase"`
	DrawIndex    int                                 `json:"drawIndex"`
	GameEndFlag  bool                                `json:"gameEndFlag"`
	LastBet      int                                 `json:"lastBet"`
	MinRaise     int                                 `json:"minRaise"`
	Ante         int                                 `json:"ante"`
	BettingLimit int                                 `json:"bettingLimit"`
	RaiseCount   int                                 `json:"raiseCount"`
	MaxBetAmount int                                 `json:"maxBetAmount"`
	RoundResults []*DeuceToSevenWebOutputResult      `json:"roundResults"`
	CpuActions   []*DeuceToSevenWebOutputCpuAction   `json:"cpuActions"`
	CpuExchanges []*DeuceToSevenWebOutputCpuExchange `json:"cpuExchanges"`
	MetaAI       *DeuceToSevenWebOutputMetaAI        `json:"metaAI,omitempty"`
	Profile      *domain.BettingHumanProfileData     `json:"profile,omitempty"`
	WebOutputBase
}

// ToConfig builds a DeuceToSevenConfig from the web input, clamping numeric
// fields to their domain-valid ranges.
func (p DeuceToSevenWebInput) ToConfig() domain.DeuceToSevenConfig {
	cfg := domain.DefaultDeuceToSevenConfig()
	cfg.CpuCount = webutil.BoundedIntPtr(p.CpuCount, domain.DeuceToSevenCpuCountMin, domain.DeuceToSevenCpuCountMax, cfg.CpuCount)
	cfg.BettingLimit = domain.BettingLimitType(webutil.BoundedIntPtr(
		p.BettingLimit, int(domain.BettingLimitFixed), int(domain.BettingLimitNoLimit), int(cfg.BettingLimit),
	))
	cfg.CpuMetaAI = p.CpuMetaAI
	return cfg
}

// DeuceToSevenWebController is the HTTP handler type for POST /deucetoseven/exec.
type DeuceToSevenWebController = GameWebController[usecase.DeuceToSevenInteractorIF, DeuceToSevenWebInput, *DeuceToSevenWebOutput]

// NewDeuceToSevenWebController / NewDeuceToSevenWebControllerWithProvider are
// the standard and session-provider-backed constructors.
var NewDeuceToSevenWebController, NewDeuceToSevenWebControllerWithProvider = webControllerPair[usecase.DeuceToSevenInteractorIF, DeuceToSevenWebInput, *DeuceToSevenWebOutput](
	newDeuceToSevenDefaultOutput, deuceToSevenDispatch,
)

func newDeuceToSevenDefaultOutput(msg string) *DeuceToSevenWebOutput {
	return &DeuceToSevenWebOutput{
		Players:       make([]*DeuceToSevenWebOutputPlayer, 0),
		SidePots:      make([]*DeuceToSevenWebOutputSidePot, 0),
		RoundResults:  make([]*DeuceToSevenWebOutputResult, 0),
		CpuActions:    make([]*DeuceToSevenWebOutputCpuAction, 0),
		CpuExchanges:  make([]*DeuceToSevenWebOutputCpuExchange, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func deuceToSevenDispatch(bc *baseController, w http.ResponseWriter, di usecase.DeuceToSevenInteractorIF, param DeuceToSevenWebInput, _ func(string) *DeuceToSevenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig(), param.Profile))
	case "e", "exchange":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, di.Exchange(indices))
	case "s", "stand":
		bc.writePresenterResponse(w, di.Stand())
	case "f", "fold":
		bc.writePresenterResponse(w, di.Action(domain.DeuceToSevenActionFold, 0, param.HumanPlayMs))
	case "ck", "check":
		bc.writePresenterResponse(w, di.Action(domain.DeuceToSevenActionCheck, 0, param.HumanPlayMs))
	case "c", "call":
		bc.writePresenterResponse(w, di.Action(domain.DeuceToSevenActionCall, 0, param.HumanPlayMs))
	case "b", "bet":
		bc.writePresenterResponse(w, di.Action(domain.DeuceToSevenActionBet, param.Amount, param.HumanPlayMs))
	case "ra", "raise":
		bc.writePresenterResponse(w, di.Action(domain.DeuceToSevenActionRaise, param.Amount, param.HumanPlayMs))
	case "a", "allin":
		bc.writePresenterResponse(w, di.Action(domain.DeuceToSevenActionAllIn, 0, param.HumanPlayMs))
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
