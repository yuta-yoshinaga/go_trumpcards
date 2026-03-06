package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
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
	Chips int                       `json:"chips"`
	Hands []*BlackJackWebOutputHand `json:"hands"`
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
	Message            string                             `json:"message"`
	MessageCode        string                             `json:"messageCode,omitempty"`
	MessageParams      map[string]string                  `json:"messageParams,omitempty"`
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
	DoubleAfterSplit   bool                               `json:"doubleAfterSplit"`
	CountingSystem     int                                `json:"countingSystem"`
	DeckPenetration    int                                `json:"deckPenetration"`
}

// BlackJackWebController ブラックジャックWebコントローラークラス
type BlackJackWebController struct {
	baseController
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
	execWithSession(&bwc.baseController, w, r, bwc.store, bwc.factory,
		func(msg string) any { return bwc.newDefaultOutput(msg) },
		nil,
		func(w rest.ResponseWriter, bji usecase.BlackJackInteractorIF, param BlackJackWebInput) bool {
			switch param.Command {
			case "r", "reset":
				if param.DealerHitsSoft17 != nil || param.CpuPlayerCount != nil || param.CountingEnabled != nil || param.DoubleAfterSplit != nil || param.CountingSystem != nil || param.DeckPenetration != nil {
					h17 := false
					if param.DealerHitsSoft17 != nil {
						h17 = *param.DealerHitsSoft17
					}
					cpuCount := 0
					if param.CpuPlayerCount != nil {
						cpuCount = *param.CpuPlayerCount
					}
					counting := false
					if param.CountingEnabled != nil {
						counting = *param.CountingEnabled
					}
					das := true
					if param.DoubleAfterSplit != nil {
						das = *param.DoubleAfterSplit
					}
					cs := 0
					if param.CountingSystem != nil {
						cs = *param.CountingSystem
					}
					dp := 0
					if param.DeckPenetration != nil {
						dp = *param.DeckPenetration
					}
					bwc.writePresenterResponse(w, bji.ResetWithConfig(h17, cpuCount, counting, das, cs, dp))
				} else {
					bwc.writePresenterResponse(w, bji.Reset())
				}
			case "h", "hit":
				bwc.writePresenterResponse(w, bji.Hit())
			case "s", "stand":
				bwc.writePresenterResponse(w, bji.Stand())
			case "b", "bet":
				ppBet := 0
				if param.PerfectPairsBet != nil {
					ppBet = *param.PerfectPairsBet
				}
				t3Bet := 0
				if param.TwentyOnePlus3Bet != nil {
					t3Bet = *param.TwentyOnePlus3Bet
				}
				bwc.writePresenterResponse(w, bji.Bet(param.Amount, ppBet, t3Bet))
			case "d", "doubledown":
				bwc.writePresenterResponse(w, bji.DoubleDown())
			case "sp", "split":
				bwc.writePresenterResponse(w, bji.Split())
			case "i", "insurance":
				bwc.writePresenterResponse(w, bji.Insurance())
			case "di", "declineinsurance":
				bwc.writePresenterResponse(w, bji.DeclineInsurance())
			case "sur", "surrender":
				bwc.writePresenterResponse(w, bji.Surrender())
			case "togglehint":
				bwc.writePresenterResponse(w, bji.ToggleHint())
			case "sd", "setdeckcount":
				bwc.writePresenterResponse(w, bji.SetDeckCount(param.Amount))
			case "togglesoft17":
				bwc.writePresenterResponse(w, bji.ToggleSoft17())
			case "togglecounting":
				bwc.writePresenterResponse(w, bji.ToggleCounting())
			case "toggledas":
				bwc.writePresenterResponse(w, bji.ToggleDAS())
			case "scs", "setcountingsystem":
				bwc.writePresenterResponse(w, bji.SetCountingSystem(param.Amount))
			case "pen", "setpenetration":
				bwc.writePresenterResponse(w, bji.SetDeckPenetration(param.Amount))
			default:
				return false
			}
			return true
		})
}

// Stop stops the background cleanup goroutine of the session store.
func (bwc *BlackJackWebController) Stop() {
	bwc.store.Stop()
}

// newDefaultOutput エラー・定型応答用のデフォルト出力を返す
func (bwc *BlackJackWebController) newDefaultOutput(msg string) *BlackJackWebOutput {
	return &BlackJackWebOutput{
		Dealer:          &BlackJackWebOutputPlayer{},
		Player:          &BlackJackWebOutputPlayer{},
		Message:         msg,
		DeckCount:       1,
		DeckPenetration: 75,
	}
}
