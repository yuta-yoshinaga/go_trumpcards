package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BlackJackSwitchWebInput ブラックジャック・スイッチWebインプット
type BlackJackSwitchWebInput struct {
	BaseWebInput
	Amount int `json:"amount,omitempty"`
}

// BlackJackSwitchWebOutputHand 1ハンド分のアウトプット
type BlackJackSwitchWebOutputHand struct {
	Cards   []*WebOutputCard `json:"cards"`
	Score   int              `json:"score"`
	Bet     int              `json:"bet"`
	Stood   bool             `json:"stood"`
	Doubled bool             `json:"doubled"`
	Busted  bool             `json:"busted"`
	IsBJ    bool             `json:"isBJ"`
	Result  int              `json:"result"`
	Payout  int              `json:"payout"`
}

// BlackJackSwitchWebOutput ブラックジャック・スイッチWebアウトプット
type BlackJackSwitchWebOutput struct {
	Hands          []*BlackJackSwitchWebOutputHand `json:"hands"`
	DealerCards    []*WebOutputCard                `json:"dealerCards"`
	DealerScore    int                             `json:"dealerScore"`
	Phase          int                             `json:"phase"`
	CurrentHandIdx int                             `json:"currentHandIdx"`
	Chips          int                             `json:"chips"`
	Switched       bool                            `json:"switched"`
	DealerPushed22 bool                            `json:"dealerPushed22"`
	OverallResult  int                             `json:"overallResult"`
	TotalPayout    int                             `json:"totalPayout"`
	WebOutputBase
}

// BlackJackSwitchWebController ブラックジャック・スイッチWebコントローラー
type BlackJackSwitchWebController = GameWebController[usecase.BlackJackSwitchInteractorIF, BlackJackSwitchWebInput, *BlackJackSwitchWebOutput]

// NewBlackJackSwitchWebController and NewBlackJackSwitchWebControllerWithProvider
// are the standard and provider-backed constructors.
var NewBlackJackSwitchWebController, NewBlackJackSwitchWebControllerWithProvider = webControllerPair[usecase.BlackJackSwitchInteractorIF, BlackJackSwitchWebInput, *BlackJackSwitchWebOutput](
	newBlackJackSwitchDefaultOutput, blackJackSwitchDispatch,
)

func newBlackJackSwitchDefaultOutput(msg string) *BlackJackSwitchWebOutput {
	return &BlackJackSwitchWebOutput{
		Hands:         make([]*BlackJackSwitchWebOutputHand, 0),
		DealerCards:   make([]*WebOutputCard, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func blackJackSwitchDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BlackJackSwitchInteractorIF, param BlackJackSwitchWebInput, _ func(string) *BlackJackSwitchWebOutput) bool {
	switch param.Command {
	case "b", "bet":
		bc.writePresenterResponse(w, bi.Bet(param.Amount))
	case "sw", "switch":
		bc.writePresenterResponse(w, bi.Switch())
	case "k", "keep":
		bc.writePresenterResponse(w, bi.Keep())
	case "h", "hit":
		bc.writePresenterResponse(w, bi.Hit())
	case "s", "stand":
		bc.writePresenterResponse(w, bi.Stand())
	case "dd", "doubledown":
		bc.writePresenterResponse(w, bi.DoubleDown())
	default:
		return dispatchResetAndLog(param.Command, bc, w, bi.Reset, bi.ActionLog)
	}
	return true
}
