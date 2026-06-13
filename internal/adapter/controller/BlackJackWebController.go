//go:build !js || !wasm || casino

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"net/http"
)

// BlackJackWebInput ブラックジャックWebインプット
type BlackJackWebInput struct {
	BaseWebInput
	Amount            int   `json:"amount,omitempty"`
	DealerHitsSoft17  *bool `json:"dealerHitsSoft17,omitempty"`
	CpuPlayerCount    *int  `json:"cpuPlayerCount,omitempty"`
	CountingEnabled   *bool `json:"countingEnabled,omitempty"`
	DoubleAfterSplit  *bool `json:"doubleAfterSplit,omitempty"`
	CountingSystem    *int  `json:"countingSystem,omitempty"`
	DeckPenetration   *int  `json:"deckPenetration,omitempty"`
	PerfectPairsBet   *int  `json:"perfectPairsBet,omitempty"`
	TwentyOnePlus3Bet *int  `json:"twentyOnePlus3Bet,omitempty"`
	HandCount         *int  `json:"handCount,omitempty"`
	SurrenderRule     *int  `json:"surrenderRule,omitempty"`
}

// BlackJackWebOutputHand ブラックジャックWebアウトプットハンド
type BlackJackWebOutputHand struct {
	Score        int              `json:"score"`
	Cards        []*WebOutputCard `json:"cards"`
	Bet          int              `json:"bet"`
	Stood        bool             `json:"stood"`
	Doubled      bool             `json:"doubled"`
	Busted       bool             `json:"busted"`
	IsBlackJack  bool             `json:"isBlackJack"`
	CanSplit     bool             `json:"canSplit"`
	Surrendered  bool             `json:"surrendered"`
	CanSurrender bool             `json:"canSurrender"`
}

// BlackJackWebOutputPlayer ブラックジャックWebアウトプットプレイヤー
type BlackJackWebOutputPlayer struct {
	Score int              `json:"score"`
	Cards []*WebOutputCard `json:"cards"`
	Chips int              `json:"chips"`
}

// BlackJackWebOutputCpuSeat CPUプレイヤー席アウトプット
type BlackJackWebOutputCpuSeat struct {
	Chips        int                       `json:"chips"`
	Hands        []*BlackJackWebOutputHand `json:"hands"`
	InsuranceBet int                       `json:"insuranceBet"`
}

// BlackJackWebOutputSideBetResult サイドベット結果アウトプット
type BlackJackWebOutputSideBetResult struct {
	BetType    int    `json:"betType"`
	ResultType int    `json:"resultType"`
	ResultName string `json:"resultName"`
	BetAmount  int    `json:"betAmount"`
	Payout     int    `json:"payout"`
}

// BlackJackWebOutput ブラックジャックWebアウトプット
type BlackJackWebOutput struct {
	Dealer             *BlackJackWebOutputPlayer          `json:"dealer"`
	Player             *BlackJackWebOutputPlayer          `json:"player"`
	Hands              []*BlackJackWebOutputHand          `json:"hands,omitempty"`
	CurrentHandIdx     int                                `json:"currentHandIdx"`
	Phase              int                                `json:"phase"`
	InsuranceBet       int                                `json:"insuranceBet"`
	InsuranceAvailable bool                               `json:"insuranceAvailable"`
	HintEnabled        bool                               `json:"hintEnabled"`
	SuggestedAction    int                                `json:"suggestedAction"`
	DeckCount          int                                `json:"deckCount"`
	DealerHitsSoft17   bool                               `json:"dealerHitsSoft17"`
	CountingEnabled    bool                               `json:"countingEnabled"`
	CpuPlayerCount     int                                `json:"cpuPlayerCount"`
	RunningCount       int                                `json:"runningCount"`
	TrueCount          float64                            `json:"trueCount"`
	CpuPlayers         []*BlackJackWebOutputCpuSeat       `json:"cpuPlayers,omitempty"`
	PerfectPairsBet    int                                `json:"perfectPairsBet"`
	TwentyOnePlus3Bet  int                                `json:"twentyOnePlus3Bet"`
	SideBetResults     []*BlackJackWebOutputSideBetResult `json:"sideBetResults,omitempty"`
	Bonuses            []string                           `json:"bonuses,omitempty"`
	DoubleAfterSplit   bool                               `json:"doubleAfterSplit"`
	CountingSystem     int                                `json:"countingSystem"`
	DeckPenetration    int                                `json:"deckPenetration"`
	MultiHandCount     int                                `json:"multiHandCount"`
	SurrenderRule      int                                `json:"surrenderRule"`
	WebOutputBase
}

// HasConfigParams reports whether any config parameter is set in the input.
func (p BlackJackWebInput) HasConfigParams() bool {
	return p.DealerHitsSoft17 != nil || p.CpuPlayerCount != nil || p.CountingEnabled != nil ||
		p.DoubleAfterSplit != nil || p.CountingSystem != nil || p.DeckPenetration != nil || p.SurrenderRule != nil
}

// ToConfig builds a BlackJackConfig from the web input pointer fields.
func (p BlackJackWebInput) ToConfig() domain.BlackJackConfig {
	return domain.BlackJackConfig{
		DealerHitsSoft17: deref(p.DealerHitsSoft17),
		CpuPlayerCount:   deref(p.CpuPlayerCount),
		CountingEnabled:  deref(p.CountingEnabled),
		DoubleAfterSplit: derefDefault(p.DoubleAfterSplit, true),
		CountingSystem:   deref(p.CountingSystem),
		DeckPenetration:  deref(p.DeckPenetration),
		SurrenderRule:    deref(p.SurrenderRule),
	}
}

// BlackJackWebController ブラックジャックWebコントローラークラス
type BlackJackWebController = GameWebController[usecase.BlackJackInteractorIF, BlackJackWebInput, *BlackJackWebOutput]

// NewBlackJackWebController and NewBlackJackWebControllerWithProvider are
// the standard and provider-backed constructors for BlackJackWebController.
var NewBlackJackWebController, NewBlackJackWebControllerWithProvider = webControllerPair[usecase.BlackJackInteractorIF, BlackJackWebInput, *BlackJackWebOutput](
	newBlackJackDefaultOutput, blackJackDispatch,
)

func newBlackJackDefaultOutput(msg string) *BlackJackWebOutput {
	return &BlackJackWebOutput{
		Dealer:          &BlackJackWebOutputPlayer{},
		Player:          &BlackJackWebOutputPlayer{},
		WebOutputBase:   WebOutputBase{Message: msg},
		DeckCount:       1,
		DeckPenetration: 75,
	}
}

func blackJackDispatch(bc *baseController, w http.ResponseWriter, bji usecase.BlackJackInteractorIF, param BlackJackWebInput, _ func(string) *BlackJackWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.HasConfigParams() {
			cfg := param.ToConfig()
			bc.writePresenterResponse(w, bji.ResetWithConfig(cfg))
		} else {
			bc.writePresenterResponse(w, bji.Reset())
		}
	case "h", "hit":
		bc.writePresenterResponse(w, bji.Hit())
	case "s", "stand":
		bc.writePresenterResponse(w, bji.Stand())
	case "b", "bet":
		ppBet := deref(param.PerfectPairsBet)
		t3Bet := deref(param.TwentyOnePlus3Bet)
		hc := deref(param.HandCount)
		bc.writePresenterResponse(w, bji.Bet(param.Amount, ppBet, t3Bet, hc))
	case "d", "doubledown":
		bc.writePresenterResponse(w, bji.DoubleDown())
	case "sp", "split":
		bc.writePresenterResponse(w, bji.Split())
	case "i", "insurance":
		bc.writePresenterResponse(w, bji.Insurance())
	case "di", "declineinsurance":
		bc.writePresenterResponse(w, bji.DeclineInsurance())
	case "sur", "surrender":
		bc.writePresenterResponse(w, bji.Surrender())
	case "es", "earlysurrender":
		bc.writePresenterResponse(w, bji.EarlySurrender())
	case "des", "declineearlysurrender":
		bc.writePresenterResponse(w, bji.DeclineEarlySurrender())
	case "ssr", "setsurrenderrule":
		bc.writePresenterResponse(w, bji.SetSurrenderRule(param.Amount))
	case "togglehint":
		bc.writePresenterResponse(w, bji.ToggleHint())
	case "sd", "setdeckcount":
		bc.writePresenterResponse(w, bji.SetDeckCount(param.Amount))
	case "togglesoft17":
		bc.writePresenterResponse(w, bji.ToggleSoft17())
	case "togglecounting":
		bc.writePresenterResponse(w, bji.ToggleCounting())
	case "toggledas":
		bc.writePresenterResponse(w, bji.ToggleDAS())
	case "scs", "setcountingsystem":
		bc.writePresenterResponse(w, bji.SetCountingSystem(param.Amount))
	case "pen", "setpenetration":
		bc.writePresenterResponse(w, bji.SetDeckPenetration(param.Amount))
	case "scc", "setcpucount":
		bc.writePresenterResponse(w, bji.SetCpuPlayerCount(param.Amount))
	default:
		return dispatchLog(param.Command, bc, w, bji.ActionLog)
	}
	return true
}
